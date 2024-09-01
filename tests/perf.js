import http from "k6/http";

// https://grafana.com/docs/k6/latest/using-k6/k6-options/reference/
export const options = {
  discardResponseBodies: true,

  thresholds: {
    http_req_failed: [{ threshold: 'rate<0.01', abortOnFail: true }], // http errors should be less than 1%, otherwise abort the test
    http_req_duration: ['p(99)<1000'], // 99% of requests should be below 1s
  },

  scenarios: {
    breaking: {
      executor: 'ramping-vus',
      stages: [
        { duration: '50s', target: 500 },
      ],
    },
  },
};

export default function() {
  http.get("http://localhost:8888/testns1/something");
}
