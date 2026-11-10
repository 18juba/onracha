import { describe, expect, it } from 'vitest';
import { discoveryRadiusKM, shouldRefreshDiscovery } from './map-discovery';

describe('map discovery request policy', () => {
	it('searches when there is no previous coverage', () => {
		expect(
			shouldRefreshDiscovery({
				centerMovementKM: 0,
				visibleRadiusKM: 4,
				previousCoverage: null
			})
		).toBe(true);
	});

	it('keeps the current results while panning within the searched area', () => {
		expect(
			shouldRefreshDiscovery({
				centerMovementKM: 8,
				visibleRadiusKM: 4,
				previousCoverage: { radiusKM: 25, visibleRadiusKM: 4 }
			})
		).toBe(false);
	});

	it('searches again after a meaningful pan reaches the search margin', () => {
		expect(
			shouldRefreshDiscovery({
				centerMovementKM: 17,
				visibleRadiusKM: 4,
				previousCoverage: { radiusKM: 25, visibleRadiusKM: 4 }
			})
		).toBe(true);
	});

	it('refreshes when zooming out exposes an area outside the previous coverage', () => {
		expect(
			shouldRefreshDiscovery({
				centerMovementKM: 0,
				visibleRadiusKM: 21,
				previousCoverage: { radiusKM: 25, visibleRadiusKM: 4 }
			})
		).toBe(true);
	});

	it('adds a buffer to the viewport and respects the API radius limits', () => {
		expect(discoveryRadiusKM(4)).toBe(25);
		expect(discoveryRadiusKM(30)).toBe(35);
		expect(discoveryRadiusKM(120)).toBe(100);
	});
});
