// Hot key: 50% of reads hammer just 10 URLs. Tests hot-spot handling —
// watch the per-node distribution panel to see which cache node absorbs it.
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const RATE = parseInt(__ENV.RATE || '1000');
const DURATION = __ENV.DURATION || '2m';

export const options = {
  scenarios: {
    hotkey: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
  },
};

export function setup() {
  const hot = [];
  const cold = [];
  for (let i = 0; i < 210; i++) {
    const res = http.post(`${BASE}/api/shorten`,
      JSON.stringify({ url: `https://example.com/hk-${i}` }),
      { headers: { 'Content-Type': 'application/json' } });
    if (res.status === 201) (i < 10 ? hot : cold).push(res.json('code'));
    sleep(0.01);
  }
  if (hot.length === 0) throw new Error('setup: could not seed any URLs — is the stack up?');
  return { hot, cold };
}

export default function (data) {
  const pool = Math.random() < 0.5 ? data.hot : data.cold;
  const code = pool[Math.floor(Math.random() * pool.length)];
  const res = http.get(`${BASE}/r/${code}`, { redirects: 0 });
  check(res, { 'redirect 302': (r) => r.status === 302 });
}
