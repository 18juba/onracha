import http from 'k6/http';
import { check, sleep } from 'k6';

const baseURL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
	vus: 1,
	duration: '30s',
	thresholds: {
		checks: ['rate>0.99'],
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<500']
	}
};

export default function () {
	const response = http.get(`${baseURL}/health`);

	check(response, {
		'health returns HTTP 200': (result) => result.status === 200,
		'health status is ok': (result) => result.json('status') === 'ok'
	});

	sleep(1);
}
