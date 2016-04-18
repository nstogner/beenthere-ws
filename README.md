# Been There (Web Service)

This web service provides a REST api for accessing user's visits to different cities/states.

### ENDPOINTS
| Method | URL | Function |
|:-------|:----|:---------|
| GET | /states/:state/cities | Getting a list of cities from in a given state |
| POST | /users/:user/visits | Adding a visit record for a given user |
| DELETE | /users/:user/visits/:visitId | Removing a visit record for a given user |
| GET | /users/:user/visits | Getting a list of visit for a given user (paginated) |
| GET | /users/:user/visits/cities | Getting a list of unique city names visited by a given user |
| GET | /users/:user/visits/states | Getting a list of unique state names visited by a given user |

### TODO
* Implement a streaming endpoint
* Improve README/documentation (include assumptions/etc)
* Create Dockerfile
* Create Kubernetes files
* Include RethinkDB schema script
* Include user auth? (Consider whether it belongs here)