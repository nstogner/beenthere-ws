# Been There (Web Service)

This web service provides a REST api to allow users to pin areas they've visited and potentially share them with other users.

### ROUTES
| Method | URL | Function |
|:-------|:----|:---------|
| GET | /states/:state/cities | Getting a list of cities from in a given state |
| POST | /users/:user/visits | Adding a visit record for a given user |
| DELETE | /users/:user/visits/:visitId | Removing a visit record for a given user |
| GET | /users/:user/visits | Getting a list of visit for a given user (paginated) |
| GET | /users/:user/visits/cities | Getting a list of unique city names visited by a given user |
| GET | /users/:user/visits/states | Getting a list of unique state names visited by a given user |
| GET | /streams/visits | Stream new visits using Server Sent Events |

**Pagination**: Pagination is done via query parameters: "start" and "limit".

### DATABASE
[RethinkDB](https://www.rethinkdb.com/) is used as the data-store. This NoSQL database was mainly chosen for it's streaming features. A social application such as this one could benefit from a feed of real-time user updates. In addition to streaming, RethinkDB aims to be very easy to administer, which reduces operational burden.

### CONSIDERATIONS
#### 1. User Authentication
User authentication probably should exist in another service. This design would have a better seperation of concerns than lumping user-access in with user-visit functionality.
#### 2. Validating States (new visit requests)
When a user adds a visit, the state is currently validated against an in-memory map of US states. This is done so that validating the given state does not require a database call and thereby slow down every new visit request.
#### 3. Validating Cities (new visit requests)
Since there are a great number of cities, the in-memory map (used with states) is less feasible. Cities could be handled in several different manners:
* Accept all cities upon new visit request, validate city offline (what would you do about invalidated visits?)
* Maintain a source of truth of cities in the database & validate on each new visit request

### TODO
* Write more test cases
* Remove hardcoded list of states, and read from db on startup
* Create Dockerfile
* Create Kubernetes files
