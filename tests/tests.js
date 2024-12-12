import http from "k6/http";

// https://jslib.k6.io/
import chai, {
  describe,
  expect,
} from "https://jslib.k6.io/k6chaijs/4.3.4.3/index.js";

chai.config.logFailures = true;

// https://grafana.com/docs/k6/latest/using-k6/k6-options/reference/
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
const hostHeader = { 'Host': 'testns1.com' }


export default function tests() {
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
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/wildcard-rewrite`, 'http://localhost:8080/any'));
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
      expect(resp.json()["url"], resp.json()["url"]).to.equal(reqs[i].url.replace(`${baseURL}/path-rewrite`, 'http://localhost:8080/any'));
    });
  });

  describe("Encoded paths are passed correctly", () => {
    const resps = [
      { method: 'GET', url: `${baseURL}/path-rewrite/hi%2Fworld/next`, params: { headers: hostHeader, }, },
      { method: 'GET', url: `${baseURL}/testns1/path-rewrite/hi%2Fworld/next` },
    ];

    http.batch(resps).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/any/hi%2Fworld/next');
    });
  });

  describe("Encoded wildcard routes are passed correctly", () => {
    const resp = http.get(`${baseURL}/wildcard-rewrite/hi%2Fworld/next`, { headers: hostHeader, });

    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/any/hi%2Fworld/next');
  });

  describe("Redirects are not automatically followed", () => {
    const resp = http.get(`${baseURL}/httpbun/redirect-to?url=https%3A%2F%2Fgoogle.com`, { headers: hostHeader, redirects: 0 });
    expect(resp.status, resp.status).to.equal(302);
  });

  describe("Only GET are allowed on /only-get", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/only-get`, params: { headers: hostHeader, }, },
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

  describe("X-Forwarded-For does not exist by default", () => {
    const resp = http.get(`${baseURL}/headers`, { responseCallback: http.setResponseCallback(http.expectedStatuses(200)) });
    expect(resp.status, resp.status).to.equal(200);
    expect(resp.headers, resp.headers).to.not.have.key('X-Forwarded-For');
  });

  describe("Accept-Encoding is gzip by default", () => {
    const resp = http.get(`${baseURL}/headers`);
    expect(resp.status, resp.status).to.equal(200);
    expect(resp.json()["headers"]["Accept-Encoding"], resp.json()["Accept-Encoding"]).to.equal('gzip');
  });

  describe("noRewritePath is handled correctly", () => {
    const resp = http.get(`${baseURL}/testns2/any/hi`);
    expect(resp.status, resp.status).to.equal(200);
  });

  describe("non-terminated paths behave correctly", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/not-terminated/hi/`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/hi/` },
      { method: 'GET', url: `${baseURL}/not-terminated/a/b/c/`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/a/b/c/` },
      { method: 'GET', url: `${baseURL}/not-terminated/a/b/c/d`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/testns1/not-terminated/a/b/c/d` },
    ];

    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/any');
    });
  });

  describe("terminated paths behave correctly", () => {
    let reqs = [
      { method: 'GET', url: `${baseURL}/terminated/hi/`, params: { headers: hostHeader } },
      { method: 'GET', url: `${baseURL}/testns1/terminated/hi/` },
    ];

    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/any');
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
      { method: 'GET', url: `${baseURL}/passthrough/get` },
      { method: 'GET', url: `${baseURL}/get`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/get');
    });

    reqs = [
      { method: 'GET', url: `${baseURL}/passthrough/any/hihi/%2F` },
      { method: 'GET', url: `${baseURL}/any/hihi/%2F`, params: { headers: { "Host": "passthrough.com" } } },
    ];
    http.batch(reqs).forEach((resp) => {
      expect(resp.status, resp.status).to.equal(200);
      expect(resp.json()["url"], resp.json()["url"]).to.equal('http://localhost:8080/any/hihi/%2F');
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

    // Test case 1: Single query parameter
    let resp = http.get(`${baseURL}/testns1/get?hi=1`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("1");

    // Test case 2: Multiple query parameters
    resp = http.get(`${baseURL}/testns1/get?hi=1&bye=2`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("1");
    expect(resp.json()["args"]["bye"], resp.json()["args"]["bye"]).to.equal("2");

    // Test case 3: Encoded query parameters
    resp = http.get(`${baseURL}/testns1/get?hi=%20hello%20world%20`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal(" hello world ");

    // Test case 4: Missing query parameters
    resp = http.get(`${baseURL}/testns1/get`);
    expect(resp.json()["args"], resp.json()["args"]).to.be.empty;

    // Test case 5: Query parameters with array values
    resp = http.get(`${baseURL}/testns1/get?hi=1&hi=2&hi=3`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.deep.equal(["1", "2", "3"]);

    // Test case 6: Query parameters with empty values
    resp = http.get(`${baseURL}/testns1/get?hi=&bye=`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("");
    expect(resp.json()["args"]["bye"], resp.json()["args"]["bye"]).to.equal("");

    // Test case 7: Query parameters with null values
    resp = http.get(`${baseURL}/testns1/get?hi=null`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("null");

    // Test case 8: Query parameters with boolean values
    resp = http.get(`${baseURL}/testns1/get?hi=true&bye=false`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("true");
    expect(resp.json()["args"]["bye"], resp.json()["args"]["bye"]).to.equal("false");

    // Test case 9: Query parameters with numeric values
    resp = http.get(`${baseURL}/testns1/get?hi=123&bye=456.789`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("123");
    expect(resp.json()["args"]["bye"], resp.json()["args"]["bye"]).to.equal("456.789");

    // Test case 10: Query parameters with mixed types
    resp = http.get(`${baseURL}/testns1/get?hi=1&bye=true&foo=null&bar=%20space%20`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("1");
    expect(resp.json()["args"]["bye"], resp.json()["args"]["bye"]).to.equal("true");
    expect(resp.json()["args"]["foo"], resp.json()["args"]["foo"]).to.equal("null");
    expect(resp.json()["args"]["bar"], resp.json()["args"]["bar"]).to.equal(" space ");

    // Test case 11: Query parameters with encoded values
    resp = http.get(`${baseURL}/testns1/get?hi=hello%20world&bye=goodbye%2Fworld`);
    expect(resp.json()["args"]["hi"], resp.json()["args"]["hi"]).to.equal("hello world");
    expect(resp.json()["url"], resp.json()["url"]).to.equal("http://localhost:8080/any?hi=hello%20world&bye=goodbye%2Fworld");
  });
}
