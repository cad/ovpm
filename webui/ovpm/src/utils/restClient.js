import axios from "axios";
import format from "string-template";

// API is an object that acts like a client for remote REST APIs.
export class API {
  // @param {String} baseURL - base url for the api, that prefixes the endpoints (e.g. https://api.example.com/v1)
  // @param {Object{}} endpoints - endpoint objects. (e.g {"login": {"path": "/items/{item_id}", method: "POST"} ... etc})
  // @param {String} authToken - this token will be used (if provided) with authenticated resources
  constructor(baseURL, endpoints, authToken) {
    this.baseURL = baseURL;
    this.endpoints = endpoints;
    this.authToken = authToken;

    // validate endpoints
    if (this.endpoints.constructor !== Object) {
      throw "endpoints must be an object: " + this.baseURL;
    } else {
      for (let endpointName in this.endpoints) {
        if (endpoints[endpointName].constructor !== Object) {
          throw "items of 'endpoints' should be endpoint object: " +
            endpoints[endpointName];
        }

        let keysShouldExist = ["path", "method"];
        for (let key in keysShouldExist) {
          if (!(keysShouldExist[key] in endpoints[endpointName])) {
            throw "endpoint object should have a key called: " +
              keysShouldExist[key];
          }
          if (
            !(
              endpoints[endpointName][keysShouldExist[key]].constructor ===
              String
            )
          ) {
            throw "endpoint object should have a key called: " +
              keysShouldExist[key];
          }
        }
      }
    }

    // validate authToken
    if (this.authToken && this.authToken.constructor !== String) {
      throw "authToken should be a string: " + this.authToken;
    }
  }

  // setAuthToken receives and stores an Authorization Token
  setAuthToken(authToken) {
    if (authToken.constructor !== String) {
      throw "authToken should be a string: " + authToken;
    }
    this.authToken = authToken;
  }

  // call receives the endpoint name, data and handlers then performs the api call.
  //
  // @param {String} endpointName - name of the endpoint which corresponds
  //   to the key of the endpoints object that is provided when constructing API object
  // @param {object} data - data to pass
  // @param {bool} performAuth - if set to true, send Authorization header with the authToken
  // @param {func} onSuccess - success handler to call
  // @param {func} onFailure - failure handler to call
  call(endpointName, data, performAuth, onSuccess, onFailure) {
    // validate endpointName
    if (!(endpointName in this.endpoints)) {
      throw endpointName + " is not available in " + this.endpoints;
    }

    let endpoint = this.endpoints[endpointName];

    let callConf = {
      method: endpoint.method,
      url: this.baseURL + format(endpoint.path, { ...data }),
      data: data
    };

    // if auth is true set auth headers
    if (performAuth) {
      callConf.headers = { Authorization: "Bearer " + this.authToken };
    }

    // actually perform the call
    axios(callConf)
      .then(onSuccess)
      .catch(onFailure);
  }
}
