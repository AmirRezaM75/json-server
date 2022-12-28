## Getting started
Clone the repository:

```bash
git clone git@github.com:AmirRezaM75/json-server.git
```

Navigate to the cloned directory:

```bash
cd json-server
```

Create ``api.json`` file with following structure:

```json
{
  "port": 3000,
  "endpoints": [
    {
      "method": "GET",
      "status": 200,
      "path": "/users",
      "jsonPath": "./users.json"
    },
    {
      "method": "POST",
      "status": 201,
      "path": "/users/:userId/comments",
      "jsonPath": "./comment.json"
    }
  ]
}
```
- ``port``: port number that server will listen on.
- ``method``: specifies HTTP request method.
Possible values are GET, POST, PUT, PATCH or DELETE.
- ``jsonPath``: path to .json file that represents contents of response.

Start JSON Server:

```bash
go run main.go
```