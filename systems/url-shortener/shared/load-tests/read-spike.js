// Read spike: 95% reads at 5x steady traffic. Tests cache effectiveness —
// if the cache is doing its job, the database barely notices.
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const RATE = parseInt(__ENV.RATE || '5000');
const DURATION = __ENV.DURATION || '1m';

export const options = {
  scenarios: {
    spike: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: 300,
      maxVUs: 1500,
    },
  },
};

export function setup() {
  const codes = [];
  for (let i = 0; i < 500; i++) {
    const res = http.post(`${BASE}/api/shorten`,
      JSON.stringify({ url: `https://example.com/spike-${i}` }),
      { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 201) codes.push(res.json('code'));
    sleep(0.005);
  }
  if (codes.length === 0) throw new Error('setup: could not seed any URLs — is the stack up?');
  return { codes };
}

export default function (data) {
  if (Math.random() < 0.95) {
    const code = data.codes[Math.floor(Math.random() * data.codes.length)];
    const res = http.get(`${BASE}/r/${code}`, { redirects: 0 });
    check(res, { 'redirect 302': (r) => r.status === 302 });
  } else {
    http.post(`${BASE}/api/shorten`,
      JSON.stringify({ url: `https://example.com/sw-${__VU}-${__ITER}` }),
      { headers: { 'Content-Type': 'application/json' } });
  }
}
