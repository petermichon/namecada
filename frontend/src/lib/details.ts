export interface DomainEvent {
	action: string;
	date: string;
}

export interface Registrar {
	name?: string;
	id?: string;
	abuseEmail?: string;
}

export interface DomainDetails {
	statuses: string[];
	events: DomainEvent[];
	registrar?: Registrar;
	nameservers: string[];
	dnssec?: boolean;
	related?: string;
}

const NOTABLE = new Set([
	'pendingdelete',
	'redemptionperiod',
	'autorenewperiod',
	'clienthold',
	'serverhold',
	'pendingrenew',
	'pendingtransfer',
	'pendingcreate',
	'inactive',
	'addperiod',
	'transferperiod'
]);

function vcardValue(vcardArray: unknown, key: string): string | null {
	if (!Array.isArray(vcardArray) || vcardArray.length < 2 || !Array.isArray(vcardArray[1]))
		return null;
	for (const item of vcardArray[1] as unknown[]) {
		if (Array.isArray(item) && item[0] === key && typeof item[3] === 'string') return item[3];
	}
	return null;
}

function publicId(publicIds: unknown, type: string): string | null {
	if (!Array.isArray(publicIds)) return null;
	for (const item of publicIds) {
		const id = item as { type?: unknown; identifier?: unknown };
		if (id?.type === type && typeof id.identifier === 'string') return id.identifier;
	}
	return null;
}

export function isNotableStatus(status: string): boolean {
	return NOTABLE.has(status.toLowerCase().replace(/[^a-z]/g, ''));
}

const PARKING_PROVIDERS: { match: RegExp; provider: string }[] = [
	{ match: /\.sedoparking\.com$/i, provider: 'Sedo' },
	{ match: /\.parkingcrew\.net$/i, provider: 'ParkingCrew' },
	{ match: /(^|\.)dan\.com$/i, provider: 'Dan.com' },
	{ match: /\.bodis\.com$/i, provider: 'Bodis' },
	{ match: /\.above\.com$/i, provider: 'Above.com' },
	{ match: /\.afternic\.com$/i, provider: 'Afternic' },
	{ match: /\.cashparking\.com$/i, provider: 'GoDaddy CashParking' },
	{ match: /\.parklogic\.com$/i, provider: 'ParkLogic' },
	{ match: /\.hugedomains\.com$/i, provider: 'HugeDomains' },
	{ match: /\.buydomains\.com$/i, provider: 'BuyDomains' },
	{ match: /\.namefind\.com$/i, provider: 'NameFind' },
	{ match: /\.dns-parking\.com$/i, provider: 'Dynadot Parking' }
];

export function parkingProvider(nameservers: string[]): string | null {
	for (const ns of nameservers) {
		const hit = PARKING_PROVIDERS.find((entry) => entry.match.test(ns));
		if (hit) return hit.provider;
	}
	return null;
}

export interface NameserverHint {
	kind: 'none' | 'parked';
	provider?: string;
}

export function nameserverHint(nameservers: string[]): NameserverHint | null {
	if (nameservers.length === 0) return { kind: 'none' };
	const provider = parkingProvider(nameservers);
	return provider ? { kind: 'parked', provider } : null;
}

export function extractDetails(raw: unknown): DomainDetails | null {
	if (!raw || typeof raw !== 'object') return null;
	const data = raw as Record<string, unknown>;
	const details: DomainDetails = { statuses: [], events: [], nameservers: [] };

	if (Array.isArray(data.status)) {
		details.statuses = data.status.filter((s): s is string => typeof s === 'string');
	}
	if (Array.isArray(data.events)) {
		for (const entry of data.events) {
			const event = entry as { eventAction?: unknown; eventDate?: unknown };
			if (typeof event?.eventAction === 'string' && typeof event?.eventDate === 'string') {
				details.events.push({ action: event.eventAction, date: event.eventDate });
			}
		}
	}
	if (Array.isArray(data.nameservers)) {
		for (const entry of data.nameservers) {
			const ns = entry as { ldhName?: unknown };
			if (typeof ns?.ldhName === 'string') details.nameservers.push(ns.ldhName);
		}
	}
	const secureDns = data.secureDNS as { delegationSigned?: unknown } | undefined;
	if (secureDns && typeof secureDns.delegationSigned === 'boolean') {
		details.dnssec = secureDns.delegationSigned;
	}
	if (Array.isArray(data.links)) {
		for (const entry of data.links) {
			const link = entry as { rel?: unknown; href?: unknown };
			if (link?.rel === 'related' && typeof link.href === 'string') details.related = link.href;
		}
	}
	if (Array.isArray(data.entities)) {
		for (const entry of data.entities) {
			const entity = entry as {
				roles?: unknown;
				vcardArray?: unknown;
				publicIds?: unknown;
				entities?: unknown;
			};
			const roles = Array.isArray(entity.roles) ? entity.roles : [];
			if (!roles.includes('registrar')) continue;
			const registrar: Registrar = {
				name: vcardValue(entity.vcardArray, 'fn') ?? undefined,
				id: publicId(entity.publicIds, 'IANA Registrar ID') ?? undefined
			};
			if (Array.isArray(entity.entities)) {
				for (const sub of entity.entities) {
					const child = sub as { roles?: unknown; vcardArray?: unknown };
					const childRoles = Array.isArray(child.roles) ? child.roles : [];
					if (childRoles.includes('abuse')) {
						registrar.abuseEmail = vcardValue(child.vcardArray, 'email') ?? undefined;
					}
				}
			}
			details.registrar = registrar;
			break;
		}
	}
	return details;
}
