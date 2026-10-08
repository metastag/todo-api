// Same as stress.js but with only GET calls
// With no sqlite write operations, p(95) roughly became 100x faster, proving the bottleneck is sqlite

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 20 },  // Warm-up to normal load
    { duration: '20s', target: 50 },  // Step-up to heavy load
    { duration: '20s', target: 100 }, // Peak stress load
    { duration: '15s', target: 100 }, // Sustain peak stress
    { duration: '15s', target: 0 },   // Recovery ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1s during stress
    http_req_failed: ['rate<0.05'],    // Error rate must remain under 5%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JSON_HEADERS = { 'Content-Type': 'application/json' };

// The setup() function runs once before any virtual users start running
export function setup() {
  const ids = [];

  // Create a few initial todos to be read during testing
  for (let i = 1; i <= 5; i++) {
    const payload = JSON.stringify({
      title: `Setup Task ${i}`,
      description: `Pre-populated todo item ${i} for read stress testing`,
    });

    const res = http.post(`${BASE_URL}/todos`, payload, { headers: JSON_HEADERS });
    if (res.status === 201) {
      try {
        const item = JSON.parse(res.body);
        if (item.id) {
          ids.push(item.id);
        }
      } catch (e) {
        // Ignore json parse error if any
      }
    }
  }

  return { ids };
}

// Virtual users execute only GET operations
export default function (data) {
  // 1. Fetch all todos
  const listRes = http.get(`${BASE_URL}/todos`);
  check(listRes, {
    'GET /todos status is 200': (r) => r.status === 200,
  });

  // 2. Fetch specific todo by ID from pre-populated items
  if (data && data.ids && data.ids.length > 0) {
    const randomId = data.ids[Math.floor(Math.random() * data.ids.length)];
    const getRes = http.get(`${BASE_URL}/todos/${randomId}`);
    check(getRes, {
      'GET /todos/:id status is 200': (r) => r.status === 200,
    });
  }

  // Short pause between iterations
  sleep(0.2);
}
