// Steady state: 80% reads / 20% writes at a constant arrival rate.
// Override with: make load-test RATE=500 DURATION=1m
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const RATE = parseInt(__ENV.RATE || '1000');
const DURATION = __ENV.DURATION || '2m';

export const options = {
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
  },
};

export function setup() {
  // Seed a working set of URLs so reads have something to hit.
  const codes = [];
  for (let i = 0; i < 200; i++) {
    const res = http.post(`${BASE}/api/shorten`,
      JSON.stringify({ url: `https://example.com/seed-${i}` }),
      { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 201) codes.push(res.json('code'));
    sleep(0.01);
  }
  if (codes.length === 0) throw new Error('setup: could not seed any URLs — is the stack up?');
  return { codes };
}

export default function (data) {
  if (Math.random() < 0.8) {
    const code = data.codes[Math.floor(Math.random() * data.codes.length)];
    const res = http.get(`${BASE}/r/${code}`, { redirects: 0 });
    check(res, { 'redirect 302': (r) => r.status === 302 });
  } else {
    const res = http.post(`${BASE}/api/shorten`,
      JSON.stringify({ url: `https://example.com/w-${__VU}-${__ITER}` }),
      { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'created 201': (r) => r.status === 201 });
  }
}
