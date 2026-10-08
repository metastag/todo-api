// This k6 test puts the API under normal load (peak 20 virtual users)


import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '15s', target: 20 }, // Ramp-up to 20 virtual users
    { duration: '30s', target: 20 }, // Stay at 20 virtual users (steady load)
    { duration: '15s', target: 0 },  // Ramp-down to 0 virtual users
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 95% of requests must complete below 300ms
    http_req_failed: ['rate<0.01'],   // Error rate must be less than 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JSON_HEADERS = { 'Content-Type': 'application/json' };

export default function () {
  // 1. Create a new todo
  const createPayload = JSON.stringify({
    title: `Load Test Task ${__VU}-${__ITER}`,
    description: 'Generated during k6 load testing',
  });

  const createRes = http.post(`${BASE_URL}/todos`, createPayload, { headers: JSON_HEADERS });
  const createOk = check(createRes, {
    'POST /todos status is 201': (r) => r.status === 201,
    'POST /todos has id': (r) => {
      try {
        return JSON.parse(r.body).id > 0;
      } catch (e) {
        return false;
      }
    },
  });

  if (!createOk) {
    sleep(1);
    return;
  }

  const todoId = JSON.parse(createRes.body).id;

  // 2. Fetch all todos
  const listRes = http.get(`${BASE_URL}/todos`);
  check(listRes, {
    'GET /todos status is 200': (r) => r.status === 200,
  });

  // 3. Get individual todo by ID
  const getRes = http.get(`${BASE_URL}/todos/${todoId}`);
  check(getRes, {
    'GET /todos/:id status is 200': (r) => r.status === 200,
  });

  // 4. Update the todo
  const updatePayload = JSON.stringify({
    title: `Updated Task ${__VU}-${__ITER}`,
    completed: true,
  });
  const updateRes = http.put(`${BASE_URL}/todos/${todoId}`, updatePayload, { headers: JSON_HEADERS });
  check(updateRes, {
    'PUT /todos/:id status is 200': (r) => r.status === 200,
  });

  // 5. Delete the todo
  const deleteRes = http.del(`${BASE_URL}/todos/${todoId}`);
  check(deleteRes, {
    'DELETE /todos/:id status is 204': (r) => r.status === 204,
  });

  sleep(1);
}
