<script lang="ts">
	import Button from './ui/Button.svelte';
	import { LocateFixed, LoaderCircle, Minus, Plus, RefreshCw } from '@lucide/svelte';
	import { api, type DiscoveredArena } from '../api';
	import {
		discoveryRadiusKM,
		shouldRefreshDiscovery,
		type DiscoveryCoverage
	} from '../map-discovery';
	import type { Theme } from '../theme';

	type Coordinates = { latitude: number; longitude: number };
	type UserLocation = {
		latitude: number;
		longitude: number;
		accuracy: number;
	};
	type Props = {
		theme: Theme;
		selectable?: boolean;
		selectedLocation?: Coordinates | null;
		onSelectLocation?: (location: Coordinates) => void;
		discoverArenas?: boolean;
		gameAvailabilityVersion?: number;
		onArenaSelect?: (arena: DiscoveredArena) => void;
		onArenasChange?: (arenas: DiscoveredArena[]) => void;
	};

	const CITY_ZOOM = 12.5;
	const MIN_AUTO_DISCOVERY_ZOOM = 9;
	const DISCOVERY_DEBOUNCE_MS = 500;
	const BRAZIL_CENTER: [number, number] = [-14.235, -51.9253];
	const TILE_URL = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';

	let {
		theme,
		selectable = false,
		selectedLocation = null,
		onSelectLocation = () => {},
		discoverArenas = false,
		gameAvailabilityVersion = 0,
		onArenaSelect = () => {},
		onArenasChange = () => {}
	}: Props = $props();
	let locationStatus = $state<'loading' | 'ready' | 'error'>('loading');
	let locationMessage = $state('');
	let mapLoadError = $state(false);
	let arenaDiscoveryStatus = $state<'idle' | 'loading' | 'ready' | 'error'>('idle');
	let discoveredArenaCount = $state(0);
	let mapControls = $state({
		requestUserLocation: () => {},
		searchArenas: () => {},
		refreshGameAvailability: () => {},
		zoomIn: () => {},
		zoomOut: () => {},
		centerOnUser: () => {}
	});

	// Sync the non-Svelte Leaflet markers after a game is booked or joined.
	$effect(() => {
		if (discoverArenas && gameAvailabilityVersion > 0) {
			mapControls.refreshGameAvailability();
		}
	});

	function attachMap(element: HTMLDivElement) {
		let disposed = false;
		let requestNumber = 0;
		let discoveryRequestNumber = 0;
		let discoveryTimer: ReturnType<typeof setTimeout> | undefined;
		let locationRequestStarted = false;
		let userLocation: UserLocation | null = null;
		let currentSelection = selectedLocation;
		let leaflet: typeof import('leaflet') | undefined;
		let map: import('leaflet').Map | undefined;
		let userMarker: import('leaflet').Marker | undefined;
		let selectionMarker: import('leaflet').Marker | undefined;
		let accuracyCircle: import('leaflet').Circle | undefined;
		let lastDiscoveryCoverage: (DiscoveryCoverage & { center: [number, number] }) | null = null;
		const arenaMarkers: { id: number; marker: import('leaflet').Marker }[] = [];
		let availableGameCounts: Record<number, number> = {};
		let discoveredArenas: DiscoveredArena[] = [];

		function arenaMarkerIcon(arena: DiscoveredArena) {
			if (!leaflet || !map) return undefined;

			const size = Math.max(34, Math.min(68, Math.round(28 + (map.getZoom() - 8) * 4)));
			const showName = map.getZoom() >= 16;
			const nameCardHeight = showName ? 34 : 0;
			const gameCount = availableGameCounts[arena.id] ?? 0;
			const photoURL = typeof arena.photo_url === 'string' ? arena.photo_url.trim() : '';
			const markerContent = document.createElement('span');
			markerContent.className = 'onracha-arena-marker';
			markerContent.style.width = `${size}px`;
			markerContent.style.height = `${size + nameCardHeight}px`;
			markerContent.setAttribute('aria-label', arena.name);

			const photo = document.createElement('span');
			photo.className = 'onracha-arena-marker__photo';
			photo.style.width = `${size}px`;
			photo.style.height = `${size}px`;

			const fallback = document.createElement('span');
			fallback.className = 'onracha-arena-marker__fallback';
			fallback.setAttribute('aria-hidden', 'true');
			fallback.textContent = '⚽';
			photo.append(fallback);

			if (photoURL) {
				const image = document.createElement('img');
				image.alt = '';
				image.decoding = 'async';
				image.dataset.photoSrc = photoURL;
				photo.append(image);
			}

			markerContent.append(photo);

			if (gameCount > 0) {
				const badge = document.createElement('span');
				badge.className = 'onracha-arena-marker__games';
				badge.setAttribute('aria-hidden', 'true');
				badge.textContent = gameCount > 9 ? '9+' : String(gameCount);
				markerContent.append(badge);
			}

			if (showName) {
				const nameCard = document.createElement('span');
				nameCard.className = 'onracha-arena-marker__label';
				nameCard.style.top = `${size + 5}px`;
				nameCard.textContent = arena.name;
				markerContent.append(nameCard);
			}

			const iconHeight = size + nameCardHeight;

			return leaflet.divIcon({
				className: 'onracha-arena-marker-icon',
				html: markerContent,
				iconSize: [size, iconHeight],
				iconAnchor: [size / 2, size / 2]
			});
		}

		function loadArenaMarkerPhoto(marker: import('leaflet').Marker) {
			const markerElement = marker.getElement();
			const image = markerElement?.querySelector<HTMLImageElement>(
				'.onracha-arena-marker__photo img'
			);
			const fallback = markerElement?.querySelector<HTMLElement>('.onracha-arena-marker__fallback');
			const photoSrc = image?.dataset.photoSrc;

			if (!image || !fallback || !photoSrc || image.dataset.photoRequested === 'true') return;

			image.dataset.photoRequested = 'true';
			image.addEventListener(
				'load',
				() => {
					fallback.hidden = true;
				},
				{ once: true }
			);
			image.addEventListener(
				'error',
				() => {
					image.remove();
					fallback.hidden = false;
				},
				{ once: true }
			);
			image.src = photoSrc;
		}

		function refreshArenaMarkerIcons() {
			if (!leaflet) return;
			for (const entry of arenaMarkers) {
				const arena = discoveredArenas.find((value) => value.id === entry.id);
				const icon = arena ? arenaMarkerIcon(arena) : undefined;
				if (icon) {
					entry.marker.setIcon(icon);
					loadArenaMarkerPhoto(entry.marker);
				}
			}
		}

		function updateArenaMarkers(arenas: DiscoveredArena[]) {
			if (!map || !leaflet) return;
			const nextIDs: number[] = [];

			for (const arena of arenas) {
				if (arena.latitude === null || arena.longitude === null) continue;
				nextIDs.push(arena.id);
				const position: [number, number] = [arena.latitude, arena.longitude];
				const icon = arenaMarkerIcon(arena);
				if (!icon) continue;

				const existingMarker = arenaMarkers.find((entry) => entry.id === arena.id)?.marker;
				if (existingMarker) {
					existingMarker.setLatLng(position);
					existingMarker.setIcon(icon);
					loadArenaMarkerPhoto(existingMarker);
					existingMarker.off('click');
					existingMarker.on('click', () => onArenaSelect(arena));
				} else {
					const marker = leaflet
						.marker(position, {
							icon,
							title: arena.name,
							alt: `Abrir detalhes da arena ${arena.name}`
						})
						.addTo(map);
					loadArenaMarkerPhoto(marker);
					marker.on('click', () => onArenaSelect(arena));
					arenaMarkers.push({ id: arena.id, marker });
				}
			}

			for (let index = arenaMarkers.length - 1; index >= 0; index -= 1) {
				if (!nextIDs.includes(arenaMarkers[index].id)) {
					arenaMarkers[index].marker.remove();
					arenaMarkers.splice(index, 1);
				}
			}
		}

		async function loadAvailableGameCounts() {
			try {
				const results = await api.listAvailableGames();
				if (disposed) return;
				availableGameCounts = {};
				for (const result of results) {
					const arenaID = result.arena.id;
					availableGameCounts[arenaID] = (availableGameCounts[arenaID] ?? 0) + 1;
				}
				refreshArenaMarkerIcons();
			} catch {
				// The arena map remains useful even when game discovery is unavailable.
			}
		}

		function visibleMapRadiusKM() {
			if (!map) return 0;
			const center = map.getCenter();
			const bounds = map.getBounds();
			const corners = [
				bounds.getNorthEast(),
				bounds.getNorthWest(),
				bounds.getSouthEast(),
				bounds.getSouthWest()
			];
			return Math.max(...corners.map((corner) => map!.distance(center, corner) / 1000));
		}

		async function loadArenasForCurrentMap(force = false) {
			if (!discoverArenas || !map) return;
			const center = map.getCenter();
			const visibleRadiusKM = visibleMapRadiusKM();
			const centerMovementKM = lastDiscoveryCoverage
				? map.distance(lastDiscoveryCoverage.center, center) / 1000
				: 0;
			if (
				!force &&
				!shouldRefreshDiscovery({
					centerMovementKM,
					visibleRadiusKM,
					previousCoverage: lastDiscoveryCoverage
				})
			) {
				return;
			}

			const currentRequest = ++discoveryRequestNumber;
			const radiusKM = discoveryRadiusKM(visibleRadiusKM);
			lastDiscoveryCoverage = {
				center: [center.lat, center.lng],
				radiusKM,
				visibleRadiusKM
			};
			arenaDiscoveryStatus = 'loading';

			try {
				const arenas = await api.discoverArenas({
					latitude: center.lat,
					longitude: center.lng,
					radius_km: radiusKM
				});
				if (disposed || currentRequest !== discoveryRequestNumber) return;
				discoveredArenas = arenas;
				discoveredArenaCount = arenas.filter(
					(arena) => arena.latitude !== null && arena.longitude !== null
				).length;
				arenaDiscoveryStatus = 'ready';
				updateArenaMarkers(arenas);
				onArenasChange(arenas);
			} catch {
				if (!disposed && currentRequest === discoveryRequestNumber) {
					arenaDiscoveryStatus = 'error';
				}
			}
		}

		function scheduleArenaSearch() {
			if (!discoverArenas || !map || map.getZoom() < MIN_AUTO_DISCOVERY_ZOOM) return;
			if (discoveryTimer) clearTimeout(discoveryTimer);
			discoveryTimer = setTimeout(() => void loadArenasForCurrentMap(), DISCOVERY_DEBOUNCE_MS);
		}

		function showSelectedLocation(location: Coordinates) {
			currentSelection = location;
			if (!map || !leaflet) return;

			const position: [number, number] = [location.latitude, location.longitude];
			if (!selectionMarker) {
				const icon = leaflet.divIcon({
					className: 'onracha-selection-marker',
					html: '<span aria-hidden="true"></span>',
					iconSize: [40, 40],
					iconAnchor: [20, 38]
				});
				selectionMarker = leaflet
					.marker(position, {
						icon,
						keyboard: false,
						title: 'Local selecionado para a arena'
					})
					.addTo(map);
			} else {
				selectionMarker.setLatLng(position);
			}
		}

		function showUserLocation(location: UserLocation) {
			userLocation = location;
			locationStatus = 'ready';

			if (!map || !leaflet) return;

			const position: [number, number] = [location.latitude, location.longitude];
			map.setView(position, CITY_ZOOM);

			if (!accuracyCircle) {
				const accentColor =
					getComputedStyle(document.documentElement).getPropertyValue('--accent').trim() ||
					'#8bcac2';
				accuracyCircle = leaflet
					.circle(position, {
						radius: Math.max(location.accuracy, 30),
						color: accentColor,
						weight: 1,
						fillColor: accentColor,
						fillOpacity: 0.12,
						interactive: false
					})
					.addTo(map);
			} else {
				accuracyCircle.setLatLng(position);
				accuracyCircle.setRadius(Math.max(location.accuracy, 30));
			}

			if (!userMarker) {
				const icon = leaflet.divIcon({
					className: 'onracha-location-marker',
					html: '<span></span>',
					iconSize: [24, 24],
					iconAnchor: [12, 12]
				});

				userMarker = leaflet.marker(position, { icon, title: 'Sua localização' }).addTo(map);
			} else {
				userMarker.setLatLng(position);
			}
		}

		function locateUser() {
			const currentRequest = ++requestNumber;
			locationRequestStarted = true;
			locationStatus = 'loading';
			locationMessage = '';

			if (!navigator.geolocation) {
				locationStatus = 'error';
				locationMessage = 'Este navegador não oferece acesso à localização.';
				return;
			}

			navigator.geolocation.getCurrentPosition(
				(position) => {
					if (disposed || currentRequest !== requestNumber) return;

					showUserLocation({
						latitude: position.coords.latitude,
						longitude: position.coords.longitude,
						accuracy: position.coords.accuracy
					});
				},
				(error) => {
					if (disposed || currentRequest !== requestNumber) return;

					locationStatus = 'error';
					locationMessage =
						error.code === 1
							? 'Permita o acesso à localização nas configurações do navegador.'
							: error.code === 3
								? 'A localização demorou demais. Tente novamente.'
								: 'Não foi possível obter sua localização. Tente novamente.';
				},
				{ enableHighAccuracy: true, maximumAge: 60_000, timeout: 15_000 }
			);
		}

		mapControls.requestUserLocation = locateUser;
		mapControls.searchArenas = () => void loadArenasForCurrentMap(true);
		mapControls.refreshGameAvailability = () => void loadAvailableGameCounts();
		mapControls.centerOnUser = () => {
			if (!userLocation) {
				locateUser();
				return;
			}

			if (!map) return;

			const position: [number, number] = [userLocation.latitude, userLocation.longitude];
			const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
			map.setView(position, CITY_ZOOM, { animate: !reduceMotion, duration: 0.6 });
		};

		async function initializeMap() {
			leaflet = await import('leaflet');
			if (disposed) return;

			map = leaflet.map(element, {
				center: BRAZIL_CENTER,
				zoom: 4,
				minZoom: 3,
				maxZoom: 19,
				zoomSnap: 0.5,
				zoomControl: false
			});

			leaflet
				.tileLayer(TILE_URL, {
					maxZoom: 19,
					updateWhenIdle: true,
					attribution:
						'&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer">OpenStreetMap contributors</a>'
				})
				.addTo(map);

			mapControls.zoomIn = () => map?.zoomIn();
			mapControls.zoomOut = () => map?.zoomOut();
			if (discoverArenas) {
				map.on('moveend', scheduleArenaSearch);
				map.on('zoomend', refreshArenaMarkerIcons);
				void loadAvailableGameCounts();
			}
			window.requestAnimationFrame(() => map?.invalidateSize());

			if (selectable) {
				map.on('click', (event: import('leaflet').LeafletMouseEvent) => {
					const location = {
						latitude: event.latlng.lat,
						longitude: event.latlng.lng
					};
					showSelectedLocation(location);
					onSelectLocation(location);
				});
			}

			if (currentSelection) showSelectedLocation(currentSelection);
			if (userLocation) showUserLocation(userLocation);
			if (discoverArenas) scheduleArenaSearch();
			if (!locationRequestStarted) locateUser();
		}

		void initializeMap().catch(() => {
			if (!disposed) mapLoadError = true;
		});

		return () => {
			disposed = true;
			requestNumber += 1;
			discoveryRequestNumber += 1;
			if (discoveryTimer) clearTimeout(discoveryTimer);
			map?.remove();
			mapControls.requestUserLocation = () => {};
			mapControls.searchArenas = () => {};
			mapControls.refreshGameAvailability = () => {};
			mapControls.zoomIn = () => {};
			mapControls.zoomOut = () => {};
			mapControls.centerOnUser = () => {};
		};
	}
</script>

<div
	class={`map-surface relative min-h-0 flex-1 overflow-hidden ${theme === 'dark' ? 'map-surface-dark' : 'map-surface-light'} ${selectable ? 'map-surface-selectable' : ''}`}
	role="region"
	aria-label={selectable
		? 'Mapa para selecionar o local da arena'
		: 'Mapa interativo da sua região'}
>
	<div class="absolute inset-0" {@attach attachMap}></div>

	{#if mapLoadError}
		<div
			class="absolute inset-x-4 top-4 z-[500] rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 text-sm text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			role="alert"
		>
			Não foi possível carregar o mapa.
		</div>
	{:else if locationStatus === 'loading' && !selectable}
		<div
			class="absolute top-4 left-4 z-[500] flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-sm text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			role="status"
			aria-live="polite"
		>
			<LoaderCircle size={16} class="animate-spin motion-reduce:animate-none" aria-hidden="true" />
			<span>Localizando você</span>
		</div>
	{:else if locationStatus === 'error' && !selectable}
		<div
			class="absolute top-4 left-4 z-[500] max-w-[calc(100%-5.5rem)] rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-3 text-sm text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			role="alert"
		>
			<p>{locationMessage}</p>
			<Button
				type="button"
				onclick={mapControls.requestUserLocation}
				variant="ghost"
				size="default"
				class="mt-2"
			>
				Tentar novamente
			</Button>
		</div>
	{:else if selectable && locationStatus === 'loading'}
		<div
			class="absolute bottom-4 left-4 z-[500] flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-xs text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			role="status"
			aria-live="polite"
		>
			<LoaderCircle size={14} class="animate-spin motion-reduce:animate-none" aria-hidden="true" />
			<span>Centralizando na sua localização</span>
		</div>
	{:else if selectable && locationStatus === 'error'}
		<div
			class="absolute bottom-4 left-4 z-[500] max-w-[calc(100%-5.5rem)] rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-3 text-xs text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			role="alert"
		>
			<p>{locationMessage}</p>
			<Button
				type="button"
				onclick={mapControls.requestUserLocation}
				variant="ghost"
				size="sm"
				class="mt-1"
			>
				Tentar localizar
			</Button>
		</div>
	{/if}

	{#if discoverArenas && locationStatus !== 'loading'}
		<div
			class="absolute bottom-4 left-4 z-[500] flex min-h-10 items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface)] px-3.5 text-xs font-medium text-[var(--text-primary)] shadow-lg backdrop-blur-xl"
			aria-live="polite"
		>
			{#if arenaDiscoveryStatus === 'loading'}
				<LoaderCircle
					size={14}
					class="animate-spin text-[var(--accent)] motion-reduce:animate-none"
					aria-hidden="true"
				/>
				<span>Procurando arenas</span>
			{:else if arenaDiscoveryStatus === 'error'}
				<span>Busca indisponível</span>
				<Button
					type="button"
					aria-label="Tentar buscar arenas novamente"
					onclick={mapControls.searchArenas}
					variant="ghost"
					size="icon"
					class=""
				>
					<RefreshCw size={14} aria-hidden="true" />
				</Button>
			{:else if arenaDiscoveryStatus === 'ready'}
				<span class="size-2 rounded-full bg-[var(--accent)]" aria-hidden="true"></span>
				<span>
					{discoveredArenaCount === 0
						? 'Nenhuma arena por aqui'
						: `${discoveredArenaCount} ${discoveredArenaCount === 1 ? 'arena' : 'arenas'} na área`}
				</span>
			{/if}
		</div>
	{/if}

	<div class="absolute top-4 right-4 z-[500] grid gap-2">
		<Button
			type="button"
			aria-label="Aumentar o zoom"
			onclick={mapControls.zoomIn}
			variant="outline"
			size="icon"
			class="shadow-lg backdrop-blur-xl"
		>
			<Plus size={19} aria-hidden="true" />
		</Button>
		<Button
			type="button"
			aria-label="Diminuir o zoom"
			onclick={mapControls.zoomOut}
			variant="outline"
			size="icon"
			class="shadow-lg backdrop-blur-xl"
		>
			<Minus size={19} aria-hidden="true" />
		</Button>
		<Button
			type="button"
			aria-label="Centralizar na minha localização"
			onclick={mapControls.centerOnUser}
			variant="outline"
			size="icon"
			class="shadow-lg backdrop-blur-xl"
		>
			<LocateFixed size={19} aria-hidden="true" />
		</Button>
	</div>
</div>

<style>
	.map-surface :global(.leaflet-container) {
		width: 100%;
		height: 100%;
		background: var(--app-background);
		font-family: inherit;
	}

	.map-surface-selectable :global(.leaflet-grab) {
		cursor: crosshair;
	}

	.map-surface :global(.leaflet-tile-pane) {
		transition: filter 240ms ease;
	}

	.map-surface-dark :global(.leaflet-tile-pane) {
		filter: grayscale(0.55) invert(1) hue-rotate(180deg) saturate(0.55) brightness(0.84);
	}

	.map-surface-light :global(.leaflet-tile-pane) {
		filter: saturate(0.72) brightness(1.02);
	}

	.map-surface :global(.leaflet-control-attribution) {
		max-width: calc(100% - 16px);
		padding: 2px 6px;
		border-top-left-radius: 8px;
		background: var(--surface);
		color: var(--text-secondary);
		backdrop-filter: blur(8px);
	}

	.map-surface :global(.leaflet-control-attribution a) {
		color: var(--text-primary);
	}

	:global(.onracha-location-marker) {
		border: 0;
		background: transparent;
	}

	:global(.onracha-location-marker span) {
		display: block;
		width: 18px;
		height: 18px;
		margin: 3px;
		border: 3px solid var(--app-background);
		border-radius: 9999px;
		background: var(--accent);
	}

	:global(.onracha-selection-marker) {
		border: 0;
		background: transparent;
	}

	:global(.onracha-selection-marker span) {
		position: relative;
		display: block;
		width: 30px;
		height: 30px;
		margin: 3px 5px 7px;
		border: 3px solid var(--app-background);
		border-radius: 50% 50% 50% 0;
		background: var(--accent);
		box-shadow: 0 3px 12px rgb(0 0 0 / 25%);
		transform: rotate(-45deg);
	}

	:global(.onracha-selection-marker span)::after {
		position: absolute;
		top: 7px;
		left: 7px;
		width: 10px;
		height: 10px;
		border-radius: 9999px;
		background: var(--app-background);
		content: '';
	}

	:global(.onracha-arena-marker-icon) {
		border: 0;
		background: transparent;
	}

	:global(.onracha-arena-marker) {
		position: relative;
		display: block;
		overflow: visible;
		animation: onracha-marker-pop 220ms ease-out both;
	}

	:global(.onracha-arena-marker__photo) {
		position: absolute;
		top: 0;
		left: 0;
		display: block;
		box-sizing: border-box;
		isolation: isolate;
		border: 3px solid var(--app-background);
		border-radius: 9999px;
		background: var(--surface-raised);
		overflow: hidden;
		box-shadow:
			0 3px 16px rgb(0 0 0 / 35%),
			0 0 0 2px var(--accent);
	}

	:global(.onracha-arena-marker img) {
		position: absolute;
		top: 0;
		left: 0;
		display: block;
		width: 100%;
		height: 100%;
		max-width: none;
		border-radius: 50%;
		object-fit: cover;
		object-position: center;
	}

	:global(.onracha-arena-marker__fallback) {
		position: absolute;
		inset: 0;
		z-index: 1;
		display: grid;
		place-items: center;
		font-size: 22px;
	}

	:global(.onracha-arena-marker__fallback[hidden]) {
		display: none;
	}

	:global(.onracha-arena-marker__games) {
		position: absolute;
		top: -5px;
		right: -6px;
		display: grid;
		min-width: 19px;
		height: 19px;
		place-items: center;
		border: 2px solid var(--app-background);
		border-radius: 9999px;
		background: var(--accent);
		padding: 0 4px;
		color: var(--accent-ink);
		font-size: 10px;
		font-weight: 800;
		line-height: 1;
	}

	:global(.onracha-arena-marker__label) {
		position: absolute;
		left: 50%;
		width: max-content;
		max-width: 180px;
		transform: translateX(-50%);
		overflow: hidden;
		border: 1px solid var(--border);
		border-radius: 12px;
		background: var(--surface);
		padding: 6px 10px;
		color: var(--text-primary);
		font-size: 11px;
		font-weight: 700;
		line-height: 16px;
		text-overflow: ellipsis;
		white-space: nowrap;
		box-shadow: 0 4px 14px rgb(0 0 0 / 24%);
		backdrop-filter: blur(12px);
	}

	@keyframes onracha-marker-pop {
		from {
			transform: scale(0.82);
		}
		to {
			transform: scale(1);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.map-surface :global(.leaflet-tile-pane) {
			transition: none;
		}

		:global(.onracha-arena-marker) {
			animation: none;
		}
	}
</style>
