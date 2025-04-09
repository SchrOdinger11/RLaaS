# Rate Limiter as a Service (RLaaS)

Rate Limiter as a Service (RLaaS) is a lightweight, containerized service that enforces API rate limiting by controlling the number of requests a client can make within a defined time window. This service decouples rate limiting logic from your main application, making it easier to manage, scale, and update independently.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)


## Overview

RLaaS provides a simple HTTP API for rate limiting. Clients can check if they are within the allowed quota, configure their own rate limits, and view current usage. The project supports multiple rate limiting algorithms (such as Fixed Window and Token Bucket) and is built with scalability in mind.

## Features

- **/check Endpoint:** Validate if a request is within allowed rate limits.
- **/configure Endpoint:** Dynamically set or update rate limiting parameters (e.g., number of requests per minute).
- **/usage Endpoint:** View current request counts and limit configurations.
- **Multiple Rate Limiting Algorithms:** Switch between fixed window and token bucket strategies.
- **Containerized Application:** Uses Docker and Docker Compose for local testing.
- **AWS Deployment Ready:** Initially deployed on AWS ECS Fargate using a CI/CD pipeline (currently terminated due to free-tier restrictions).

## Tech Stack

- **Programming Language:** Go
- **Data Store:** Redis
- **Containerization:** Docker & Docker Compose
- **Cloud Platform:** AWS (ECR, ECS Fargate, CodeBuild, CodePipeline - deployment terminated on free tier)

## Architecture

RLaaS is built as a two-container application:
- **HTTP Service:** A Go application handling API endpoints.
- **Redis:** An in-memory datastore for managing rate limit counters (or token buckets).

Locally, Docker Compose orchestrates these two containers. In production, the containers were originally deployed on AWS using ECS Fargate with an automated CI/CD pipeline for continuous integration and deployment.
