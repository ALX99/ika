import http from "k6/http";
import encoding from "k6/encoding";

// https://jslib.k6.io/
import chai, {
  describe,
  expect,
} from "https://jslib.k6.io/k6chaijs/4.3.4.3/index.js";

chai.config.logFailures = true;

// https://grafana.com/docs/k6/latest/using-k6/k6-options/reference/
export let options = {
  thresholds: {
    checks: ["rate == 1.00"],
    http_req_failed: ["rate == 0.00"],
  },
  iterations: 1,
  batch: 100,
  throw: true
};

const baseURL = 'http://localhost:8888';
const hostHeader = { 'Host': 'testns1.com' }


export default function testSuite() {
  describe("It is possible to send a request with any method", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      //{ method: 'HEAD', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'POST', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'PUT', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'PATCH', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'DELETE', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'CONNECT', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'OPTIONS', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
      { method: 'TRACE', url: `${baseURL}/any`, params: { headers: hostHeader, }, },
    ];

    http.batch(reqs).forEach((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["method"], resp.json()["method"]).to.equal(reqs[i].method);
    });
  });

  describe("Namespaced paths are created", () => {
    const resp = http.get(`${baseURL}/testns1/get`);

    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["method"], resp.json()["method"]).to.equal('GET');
  });

  describe("It is possible to wildcard rewrite the backend path", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/wildcard-rewrite/1`, params: { headers: hostHeader, }, },
      //{ method: 'GET', url: `${baseURL}/wildcard-rewrite/a/efgh`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/wildcard-rewrite/a/huhh`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/wildcard-rewrite/a/huhh?abc=lol&x=b`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/wildcard-rewrite/a/huhh?abc=魚&x=は`, params: { headers: hostHeader, }, },
    ];

    http.batch(reqs).forEach((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/wildcard-rewrite`, 'http://httpbun-local/any'));
    });
  });

  describe("It is possible to rewrite the backend path", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/path-rewrite/a/efgh`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh?abc=lol&x=b`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh?abc=魚&x=は`, params: { headers: hostHeader, }, },
    ];

    http.batch(reqs).forEach((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/path-rewrite`, 'http://httpbun-local/any'));
    });
  });

  describe("Encoded paths are passed correctly", () => {
    const resp = http.get(`${baseURL}/path-rewrite/hi%2Fworld/next`, { headers: hostHeader, });

    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["url"], resp.json()["url"]).to.equal('http://httpbun-local/any/hi%2Fworld/next');
  });

  describe("Encoded wildcard routes are passed correctly", () => {
    const resp = http.get(`${baseURL}/wildcard-rewrite/hi%2Fworld/next`, { headers: hostHeader, });

    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["url"], resp.json()["url"]).to.equal('http://httpbun-local/any/hi%2Fworld/next');
  });

}
