// This k6 test puts the API under much higher load than normal (peak 100 virtual users)

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

export default function () {
  // 1. Create a todo under stress
  const createPayload = JSON.stringify({
    title: `Stress Task VU-${__VU} Iter-${__ITER}`,
    description: 'Stress testing workload',
  });

  const createRes = http.post(`${BASE_URL}/todos`, createPayload, { headers: JSON_HEADERS });
  const createOk = check(createRes, {
    'POST /todos status is 201': (r) => r.status === 201,
  });

  if (!createOk) {
    sleep(0.5);
    return;
  }

  let todoId;
  try {
    todoId = JSON.parse(createRes.body).id;
  } catch (e) {
    sleep(0.5);
    return;
  }

  // 2. Fetch list under stress
  const listRes = http.get(`${BASE_URL}/todos`);
  check(listRes, {
    'GET /todos status is 200': (r) => r.status === 200,
  });

  // 3. Update todo
  const updatePayload = JSON.stringify({
    title: `Stress Updated VU-${__VU}`,
    completed: true,
  });
  const updateRes = http.put(`${BASE_URL}/todos/${todoId}`, updatePayload, { headers: JSON_HEADERS });
  check(updateRes, {
    'PUT /todos/:id status is 200': (r) => r.status === 200,
  });

  // 4. Delete todo
  const deleteRes = http.del(`${BASE_URL}/todos/${todoId}`);
  check(deleteRes, {
    'DELETE /todos/:id status is 204': (r) => r.status === 204,
  });

  // Short pause between iterations to simulate aggressive client cycles
  sleep(0.2);
}
