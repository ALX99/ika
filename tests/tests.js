import http from "k6/http";
import chai, { describe, expect } from "https://jslib.k6.io/k6chaijs/4.3.4.3/index.js";

chai.config.logFailures = true;

export const options = {
  thresholds: {
    checks: ["rate == 1.00"],
    http_req_failed: ["rate == 0.00"],
  },
  iterations: 1,
  batch: 100,
  throw: true
};

const baseURL = 'http://localhost:8888';
const testNS1URL = 'http://testns1.com';
const hostHeader = { 'Host': 'testns1.com' };
const httpbunHost = `${__ENV.HTTPBUN_HOST}`;

export default function tests() {
  describe("It is possible to send a request with any method", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'POST', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'PUT', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'PATCH', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'DELETE', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'CONNECT', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'OPTIONS', url: `${baseURL}/any`, params: { headers: hostHeader } },
      { method: 'TRACE', url: `${baseURL}/any`, params: { headers: hostHeader } },
    ];
    http.batch(reqs).map((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["method"], resp.json()["method"]).to.equal(reqs[i].method);
      return resp;
    });
  });

  describe("Namespaced paths are created", () => {
    const resp = http.get(`${baseURL}/testns1/get`);
    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["method"], resp.json()["method"]).to.equal('GET');
  });

  describe("It is possible to wildcard rewrite the backend path", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/httpbun/any/1`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/httpbun/any/a/huhh`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/httpbun/any/a/huhh?abc=lol&x=b`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/httpbun/any/a/huhh?abc=魚&x=は`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/httpbun/any/slash%2Fshould-bekept/next`, params: { headers: hostHeader } },
    ];
    http.batch(reqs).forEach((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/httpbun`, `${httpbunHost}`));
    });
  });


  describe("Retain host header test", () => {
    const resp = http.get(`${baseURL}/retain-host`, { headers: hostHeader });
    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["headers"]["Host"], resp.json()["headers"]["Host"]).to.equal(`testns1.com`);
  });

  describe("It is possible to rewrite the backend path", () => {
    const reqs = [
      { method: 'GET', url: `${baseURL}/path-rewrite/a/efgh`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh?abc=lol&x=b`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/path-rewrite/a/huhh?abc=魚&x=は`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/path-rewrite/slash%2Fshould-bekept/next`, params: { headers: hostHeader } },
    ];
    http.batch(reqs).forEach((resp, i) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/path-rewrite`, `${httpbunHost}/any`));
    });
  });

  describe("Redirects are not automatically followed", () => {
    const resp = http.get(`${baseURL}/httpbun/redirect-to?url=https%3A%2F%2Fgoogle.com`, { headers: hostHeader, redirects: 0 });
    expect(resp.status, resp.status).to.equal(302);
  });

  describe("Only GET are allowed on /only-get", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/only-get`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/testns1/only-get` },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
    });

    reqs = [
      { method: 'POST', url: `${baseURL}/only-get`, params: { headers: hostHeader, responseCallback: http.setResponseCallback(http.expectedStatuses(405)) } },
      { method: 'POST', url: `${baseURL}/testns1/only-get`, params: { responseCallback: http.setResponseCallback(http.expectedStatuses(405)) } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(405);
    });
  });

  describe("Header tests", () => {
    const resp = http.get(`${baseURL}/headers`, { responseCallback: http.setResponseCallback(http.expectedStatuses(200)) });
    describe("X-Forwarded-For does not exist by default", () => {
      expect(resp.headers, resp.headers).to.not.have.key('X-Forwarded-For');
    });
  });

  describe("non-terminated paths behave correctly", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/hi/` },
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/a/b/c/` },
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/a/b/c/d` },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(`${httpbunHost}/any`);
    });
  });

  describe("terminated paths behave correctly", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/terminated/hi/`, params: { headers: hostHeader } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(`${httpbunHost}/any`);
    });

    reqs = [
      { method: 'GET', url: `${baseURL}/terminated/hi/a/b/c/`, params: { headers: hostHeader, responseCallback: http.setResponseCallback(http.expectedStatuses(404)) } },
      { method: 'GET', url: `${baseURL}/testns1/terminated/hi/a/b/c/`, params: { responseCallback: http.setResponseCallback(http.expectedStatuses(404)) } },
      { method: 'GET', url: `${baseURL}/terminated/hi/a/b/c/d`, params: { headers: hostHeader, responseCallback: http.setResponseCallback(http.expectedStatuses(404)) } },
      { method: 'GET', url: `${baseURL}/testns1/terminated/hi/a/b/c/d`, params: { responseCallback: http.setResponseCallback(http.expectedStatuses(404)) } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(404);
    });
  });

  describe("passthrough tests", () => {
    http.setResponseCallback(http.expectedStatuses(200));
    let reqs = [
      { method: 'GET', url: `${baseURL}/passthrough`, params: { redirects: 0 } },
      { method: 'GET', url: `${baseURL}`, params: { headers: { "Host": "passthrough.com" }, redirects: 0 } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
    });

    reqs = [
      { method: 'GET', url: `${baseURL}/passthrough/` },
      { method: 'GET', url: `${baseURL}/`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
    });

    reqs = [
      { method: 'GET', url: `${baseURL}/get`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(`${httpbunHost}/get`);
    });

    reqs = [
      { method: 'GET', url: `${baseURL}/any/hihi/%2F`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal(`${httpbunHost}/any/hihi/%2F`);
    });

    http.setResponseCallback(http.expectedStatuses(404));
    reqs = [
      { method: 'GET', url: `${baseURL}/passthrough%2F` },
      { method: 'GET', url: `${baseURL}/%2F`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(404);
    });
  });

  describe("Query params are passed correctly", () => {
    http.setResponseCallback(http.expectedStatuses(200));

    const testCases = [
      { url: `${baseURL}/testns1/get?hi=1`, expected: { hi: "1" } },
      { url: `${baseURL}/testns1/get?hi=1&bye=2`, expected: { hi: "1", bye: "2" } },
      { url: `${baseURL}/testns1/get?hi=%20hello%20world%20`, expected: { hi: " hello world " } },
      { url: `${baseURL}/testns1/get`, expected: {} },
      { url: `${baseURL}/testns1/get?hi=1&hi=2&hi=3`, expected: { hi: ["1", "2", "3"] } },
      { url: `${baseURL}/testns1/get?hi=&bye=`, expected: { hi: "", bye: "" } },
      { url: `${baseURL}/testns1/get?hi=null`, expected: { hi: "null" } },
      { url: `${baseURL}/testns1/get?hi=true&bye=false`, expected: { hi: "true", bye: "false" } },
      { url: `${baseURL}/testns1/get?hi=123&bye=456.789`, expected: { hi: "123", bye: "456.789" } },
      { url: `${baseURL}/testns1/get?hi=1&bye=true&foo=null&bar=%20space%20`, expected: { hi: "1", bye: "true", foo: "null", bar: " space " } },
      { url: `${baseURL}/testns1/get?hi=hello%20world&bye=goodbye%2Fworld`, expected: { hi: "hello world", bye: "goodbye/world" } },
    ];

    testCases.forEach(({ url, expected }) => {
      const resp = http.get(url);
      expect(resp.status, resp.status).to.equal(200);
      Object.keys(expected).forEach(key => {
        expect(resp.json()["args"][key], resp.json()["args"][key]).to.deep.equal(expected[key]);
      });
    });
  });
}
