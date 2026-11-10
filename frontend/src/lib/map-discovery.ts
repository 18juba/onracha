export type DiscoveryCoverage = {
	radiusKM: number;
	visibleRadiusKM: number;
};

const DISCOVERY_RADIUS_BUFFER_KM = 5;
const MIN_DISCOVERY_SHIFT_KM = 15;
const MIN_DISCOVERY_RADIUS_KM = 25;
const MAX_DISCOVERY_RADIUS_KM = 100;

export function discoveryRadiusKM(visibleRadiusKM: number): number {
	return Math.min(
		MAX_DISCOVERY_RADIUS_KM,
		Math.max(MIN_DISCOVERY_RADIUS_KM, Math.ceil(visibleRadiusKM + DISCOVERY_RADIUS_BUFFER_KM))
	);
}

export function shouldRefreshDiscovery(input: {
	centerMovementKM: number;
	visibleRadiusKM: number;
	previousCoverage: DiscoveryCoverage | null;
}): boolean {
	const { centerMovementKM, visibleRadiusKM, previousCoverage } = input;
	if (!previousCoverage) return true;

	const expandedPastCoverage =
		visibleRadiusKM > previousCoverage.visibleRadiusKM + DISCOVERY_RADIUS_BUFFER_KM &&
		visibleRadiusKM + DISCOVERY_RADIUS_BUFFER_KM > previousCoverage.radiusKM;
	if (expandedPastCoverage) return true;

	const minimumMovementKM = Math.max(MIN_DISCOVERY_SHIFT_KM, previousCoverage.radiusKM / 2);
	if (centerMovementKM < minimumMovementKM) return false;

	return (
		centerMovementKM + visibleRadiusKM + DISCOVERY_RADIUS_BUFFER_KM > previousCoverage.radiusKM
	);
}
