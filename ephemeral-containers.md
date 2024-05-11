# Kube API Server Ephemeral Container API
Cannot patch pods using `kubectl edit` or `kubectl patch`. Only using `kubectl debug` or below API call for [pod ephemeral container subresource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#ephemeralcontainer-v1-core)

[API Server Swagger Source](https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json)

```json
"/api/v1/namespaces/{namespace}/pods/{name}/ephemeralcontainers": {
      "get": {
        "consumes": [
          "*/*"
        ],
        "description": "read ephemeralcontainers of the specified Pod",
        "operationId": "readCoreV1NamespacedPodEphemeralcontainers",
        "produces": [
          "application/json",
          "application/yaml",
          "application/vnd.kubernetes.protobuf"
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/io.k8s.api.core.v1.Pod"
            }
          },
          "401": {
            "description": "Unauthorized"
          }
        },
        "schemes": [
          "https"
        ],
        "tags": [
          "core_v1"
        ],
        "x-kubernetes-action": "get",
        "x-kubernetes-group-version-kind": {
          "group": "",
          "kind": "Pod",
          "version": "v1"
        }
      },
      "parameters": [
        {
          "description": "name of the Pod",
          "in": "path",
          "name": "name",
          "required": true,
          "type": "string",
          "uniqueItems": true
        },
        {
          "$ref": "#/parameters/namespace-vgWSWtn3"
        },
        {
          "$ref": "#/parameters/pretty-tJGM1-ng"
        }
      ],
      "patch": {
        "consumes": [
          "application/json-patch+json",
          "application/merge-patch+json",
          "application/strategic-merge-patch+json",
          "application/apply-patch+yaml"
        ],
        "description": "partially update ephemeralcontainers of the specified Pod",
        "operationId": "patchCoreV1NamespacedPodEphemeralcontainers",
        "parameters": [
          {
            "$ref": "#/parameters/body-78PwaGsr"
          },
          {
            "description": "When present, indicates that modifications should not be persisted. An invalid or unrecognized dryRun directive will result in an error response and no further processing of the request. Valid values are: - All: all dry run stages will be processed",
            "in": "query",
            "name": "dryRun",
            "type": "string",
            "uniqueItems": true
          },
          {
            "$ref": "#/parameters/fieldManager-7c6nTn1T"
          },
          {
            "description": "fieldValidation instructs the server on how to handle objects in the request (POST/PUT/PATCH) containing unknown or duplicate fields. Valid values are: - Ignore: This will ignore any unknown fields that are silently dropped from the object, and will ignore all but the last duplicate field that the decoder encounters. This is the default behavior prior to v1.23. - Warn: This will send a warning via the standard warning response header for each unknown field that is dropped from the object, and for each duplicate field that is encountered. The request will still succeed if there are no other errors, and will only persist the last of any duplicate fields. This is the default in v1.23+ - Strict: This will fail the request with a BadRequest error if any unknown fields would be dropped from the object, or if any duplicate fields are present. The error returned from the server will contain all unknown and duplicate fields encountered.",
            "in": "query",
            "name": "fieldValidation",
            "type": "string",
            "uniqueItems": true
          },
          {
            "$ref": "#/parameters/force-tOGGb0Yi"
          }
        ],
        "produces": [
          "application/json",
          "application/yaml",
          "application/vnd.kubernetes.protobuf"
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/io.k8s.api.core.v1.Pod"
            }
          },
          "201": {
            "description": "Created",
            "schema": {
              "$ref": "#/definitions/io.k8s.api.core.v1.Pod"
            }
          },
          "401": {
            "description": "Unauthorized"
          }
        },
        "schemes": [
          "https"
        ],
        "tags": [
          "core_v1"
        ],
        "x-kubernetes-action": "patch",
        "x-kubernetes-group-version-kind": {
          "group": "",
          "kind": "Pod",
          "version": "v1"
        }
      },
```
