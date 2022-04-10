
# Restful API GO with Gin, Gorm, Redis, Dockerized and Kubernetes

A boilerplate/starter project for quickly building RESTful APIs using Golang, Gin, Gorm, Redis, Dockerized and Kubernetes.


## Prerequisite

- Install Docker (https://www.docker.com/get-started/)
- Install Minikube (https://minikube.sigs.k8s.io/docs/start/)
    
## Features

- Dockerized
- CRUD API
- Unit tests in local and container


## Run Locally

Clone the project

```bash
  git clone https://github.com/aldytmra/restful-api-go.git
```

Go to the project directory

```bash
  cd my-project
```

Install dependencies

```bash
  go get ./...
```
Run all test via docker (Start the minikube)

```bash
  docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

Run API via docker 

```bash
  docker-compose up
```

Call API via postman

```bash
  http://127.0.0.1:8080/users
```


Run API via kubernetes (Start the minikube)

```bash
  minikube start
  minikube service restful-api-go --url
```

Call API via postman

```bash
  http://127.0.0.1:58563/users //example
```




