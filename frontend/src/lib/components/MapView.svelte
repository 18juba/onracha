<script lang="ts">
	import { LocateFixed, LoaderCircle, Minus, Plus } from '@lucide/svelte';
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
	};

	const CITY_ZOOM = 12.5;
	const BRAZIL_CENTER: [number, number] = [-14.235, -51.9253];
	const TILE_URL = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';

	let {
		theme,
		selectable = false,
		selectedLocation = null,
		onSelectLocation = () => {}
	}: Props = $props();
	let locationStatus = $state<'loading' | 'ready' | 'error'>('loading');
	let locationMessage = $state('');
	let mapLoadError = $state(false);
	let mapControls = $state({
		requestUserLocation: () => {},
		zoomIn: () => {},
		zoomOut: () => {},
		centerOnUser: () => {}
	});

	function attachMap(element: HTMLDivElement) {
		let disposed = false;
		let requestNumber = 0;
		let locationRequestStarted = false;
		let userLocation: UserLocation | null = null;
		let currentSelection = selectedLocation;
		let leaflet: typeof import('leaflet') | undefined;
		let map: import('leaflet').Map | undefined;
		let userMarker: import('leaflet').Marker | undefined;
		let selectionMarker: import('leaflet').Marker | undefined;
		let accuracyCircle: import('leaflet').Circle | undefined;

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
				accuracyCircle = leaflet
					.circle(position, {
						radius: Math.max(location.accuracy, 30),
						color: 'rgba(9, 176, 244, 0.6)',
						weight: 1,
						fillColor: '#09b0f4',
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
			if (!locationRequestStarted) locateUser();
		}

		void initializeMap().catch(() => {
			if (!disposed) mapLoadError = true;
		});

		return () => {
			disposed = true;
			requestNumber += 1;
			map?.remove();
			mapControls.requestUserLocation = () => {};
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
			<button
				class="mt-2 min-h-11 rounded-xl px-3 font-medium text-[var(--accent)] transition-colors hover:bg-[var(--focus-ring)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				onclick={mapControls.requestUserLocation}
			>
				Tentar novamente
			</button>
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
			<button
				class="mt-1 min-h-9 font-semibold text-[var(--accent)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
				type="button"
				onclick={mapControls.requestUserLocation}
			>
				Tentar localizar
			</button>
		</div>
	{/if}

	<div class="absolute top-4 right-4 z-[500] grid gap-2">
		<button
			class="grid size-11 place-items-center rounded-2xl border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] shadow-lg backdrop-blur-xl transition-colors hover:bg-[var(--surface-raised)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
			type="button"
			aria-label="Aumentar o zoom"
			onclick={mapControls.zoomIn}
		>
			<Plus size={19} aria-hidden="true" />
		</button>
		<button
			class="grid size-11 place-items-center rounded-2xl border border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] shadow-lg backdrop-blur-xl transition-colors hover:bg-[var(--surface-raised)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
			type="button"
			aria-label="Diminuir o zoom"
			onclick={mapControls.zoomOut}
		>
			<Minus size={19} aria-hidden="true" />
		</button>
		<button
			class="grid size-11 place-items-center rounded-2xl border border-[var(--border)] bg-[var(--surface)] text-[var(--accent)] shadow-lg backdrop-blur-xl transition-colors hover:bg-[var(--focus-ring)] focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] focus-visible:outline-none"
			type="button"
			aria-label="Centralizar na minha localização"
			onclick={mapControls.centerOnUser}
		>
			<LocateFixed size={19} aria-hidden="true" />
		</button>
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
		box-shadow: 0 0 0 5px rgb(9 176 244 / 20%);
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

	@media (prefers-reduced-motion: reduce) {
		.map-surface :global(.leaflet-tile-pane) {
			transition: none;
		}
	}
</style>
