# NASA largest picture API

This is simple API that finds largest Mars picture by sol. Sol is the
time period that NASA use to divide mission days.

## Task scope and expectations

- User should be able to supply POST request command with sol body and receive command id
- Using given command id user can create GET request and receive the largest picture for given sol
- When supplying command calculation of largest picture should be async


## Functional requirements
- There should be an endpoint <code>../mars/pictures/largest/command</code> that collects commands
- User should receive an instant response that his command accepted and calculation of largest picture has started
- Parallel processing of pictures should take place so that response time during <code>../mars/pictures/largest/command/{commandId}</code> would be quick
- Largest pictures should be stored in PostgreSQL, so that already visited sol would not cause another largest picture calculation


## Technical requirements
- The server should expose two REST API endpoints for command collection and picture return
- If user requests picture while it is being calculated this should be handled without causing duplicate calculation
- Once command received, message should be sent to the RabbitMQ broker for the command processing
- if user supplies sol for which calculation is happening already then server should not initiate the largest picture calculation again 


## How to run
- `make dc` runs docker-compose with the app container on port 8080 for you.
- `make test` runs the tests
- `make run` runs the app locally on port 8080 without docker.
- `make lint` runs the linter


## Solution notes
- clean architecture (handler->service->repository)
- docker compose + Makefile included
- PostgreSQL migrations included
- Postman collection included