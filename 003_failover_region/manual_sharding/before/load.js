import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export default function () {
  const headers = {
    'Content-Type': 'application/json'
  };

  const customer = createCustomer();
  const res = http.post(
    `http://localhost:3000/customers`,
    JSON.stringify(customer),
    { headers: headers}
  );

  check(res, { 'status was 200': (r) => r.status == 200 });
  sleep(1);
}

function createCustomer() {
  return {
    id: uuidv4(),
    email: `${(Math.random() + 1).toString(36).substring(2)}@gmail.com`
  };
}
