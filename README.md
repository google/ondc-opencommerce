# ONDC Open Commerce

ONDC aims to democratize access to commerce, by decoupling buyers, sellers and other stakeholders in the commerce ecosystem and making them interoperable through a common network. ONDC Open Commerce is an open source repository that provides code to the network participants in a way that they find it easy to integrate into the ONDC network, perform requisite commerce operations and never have to worry about integration, scalability and security.

This solution is built on [ONDC v1.2.0](https://docs.google.com/document/d/1aRzox3_Dq0Q_SicIaKegdU7FpM5q8R1rjrA6vi8qF0E/edit) and leverages the following technologies:

- [Golang](https://go.dev/)
- [Bazel](https://bazel.build/)
- [Terraform](https://www.terraform.io/)

Value Proposition
- Supports Traffic Shaping to protect retail backend systems - Allows customers to configure the shape the traffic to their systems
- Extensible - Partners can fork the services to add value additions like rule engine to filter the requests, analytics & search catalogs
- Portable Interface for existing ONDC participants - Accelerator’s interface is compatible with ONDC specification allowing participants to port to the service faster
- Compatible with ONDC Latest Specification

## New to ONDC

In case you are new to the ONDC network, we recommend you go through the following documents, which should give you an overview of steps and development for connecting to the ONDC network.

- [ONDC Development Guide](https://ondc-issue-logging-cohort1.atlassian.net/wiki/spaces/TG/pages/35160065/ONDC+Developer+Guide)
- [ONDC Integration Guide](https://docs.google.com/presentation/d/1HPRXk3lVYKmyAFcApgukZuwHhIZ_VlqR/edit#slide=id.g142ae05b320_0_0)

## Service Overview

This section will describe the services available in this repository. All services are available in a form of source code, Docker images and Terraform modules for deploying them to GCP.

#### Onboarding / Registration Service
It implements `/on_subscribe` API and `/ondc-site-verification.html`, which both are required for onboarding to the ONDC network in `pre-production` and `production` environments.

#### Key management Service
It implements key generation and key rotation for the signing key and the encryption key.

#### Core API Adapter
It provides middleware components that sit between your open-commerce applications and the ONDC network. The middleware provides the following features.
- sign and verify the authentication header.
- validate incoming request payload based on the OpenAPI specification.
- store transaction logs in the Spanner database.
- convert an asynchronous communication into a synchronous communication.

The adapter consists of 2 modules
1. Buyer Platform for buyer app
2. Seller Platform for seller app


## Requirements

This solution is only applicable for ONDC network participants and open-commerce applications with the following properties

- Use Retail Domain.
- Use [API Contract v1.2.0](https://docs.google.com/document/d/1aRzox3_Dq0Q_SicIaKegdU7FpM5q8R1rjrA6vi8qF0E/edit).
- Use [OpenAPI Specification v1.0.31](https://app.swaggerhub.com/apis/ONDC/ONDC-Protocol-Retail/1.0.31#/).
- Act as a buyer or non-msn seller role in the network. You can find out more about roles in ONDC, see [Role Selection](https://docs.google.com/presentation/d/1HPRXk3lVYKmyAFcApgukZuwHhIZ_VlqR/edit#slide=id.g2762262756f_71_128)

## Getting Started

### Prerequisites

The following is needed for building and deploying the services.

- [Golang](https://go.dev/doc/install) ~> 1.20.7
- [Bazel](https://bazel.build/install) ~> 6.3.1
- [Terraform](https://developer.hashicorp.com/terraform/downloads) ~> 0.13

(Optional) Once `bazel` is installed, you can make sure that all required dependencies are up-to-date by running
`bazel run //:gazelle-update-repos && bazel run //:gazelle` 


### Building Docker Images

Docker images of ONDC Open Commerce services are required to be stored on your Artifact Registry. Terraform scripts will access your Artifact Registry for provision of the ONDC Open Commerce service. To create an Docker repository on Artifact Registry, see [Create standard repositories](https://cloud.google.com/artifact-registry/docs/repositories/create-repos#docker)

We utilize `bazel` for building Docker images from Go source code and publishing them to the Artifact Registry.

Here is an example command to build and publish a Docker image of a specific service.
```shell
bazel run //docker/publish/onboarding:server_image_pusher_onboarding --define DOCKER_REGISTRY="asia-southeast1-docker.pkg.dev" --define DOCKER_REPOSITORY="project-id/repo-name"
```
For simple usage, shell scripts for building docker images are provided under the directory [docker/scripts/](docker/scripts/).
```
docker/scripts
├── publish_all.sh
├── publish_buyer.sh
├── publish_keyrotation.sh
├── publish_mockup.sh
├── publish_onboarding.sh
└── publish_seller.sh
```

You can build and publish all Docker images for all modules by running the following command:
```shell
# Change these variables to match your Docker repo
DOCKER_REGISTRY="asia-southeast1-docker.pkg.dev"
DOCKER_REPOSITORY="project-id/repo-name"

./docker/scripts/publish_all.sh $DOCKER_REGISTRY $DOCKER_REPOSITORY
```

You can build and publish all Docker images for a specific modules by running the following command:
```shell
# Change these variables to match your Docker repo
DOCKER_REGISTRY="asia-southeast1-docker.pkg.dev"
DOCKER_REPOSITORY="project-id/repo-name"

# Example: onboarding module
./docker/scripts/publish_onboarding.sh $DOCKER_REGISTRY $DOCKER_REPOSITORY
```

### Terraform Deployment

We provide a Terraform module for each service to help you deploy the services to GCP. If you are new to Terraform on Google Cloud, see the this [guide](https://cloud.google.com/docs/terraform/maturity)

There are example usages provided in the [terraform/examples/sample/](terraform/examples/sample/) folder. This should give you an example of how to utilize each module and connect them to work together.

#### Modules
This is a list of Terraform modules we provide.
- [Buyer Module](./terraform/modules/buyer/) - Core API Adapter for Buyer app
- [Seller Module](./terraform/modules/seller/) - Core API Adapter for Seller app
- [Key Rotation Module](./terraform/modules/key-rotation/) - Key Rotation service
- [Onboarding Module](./terraform/modules/onboarding/) - Onboarding service
- [Load Balancer](./terraform/modules/helpers/loadbalancer/) - Helper Module to deploy Application Load Balancer

#### Note
- Your patience is appreciated as certificate provisioning can take a long time to be completed (> 30 minutes).

- The Anthos Service Mesh must be enabled ONLY ONCE. You are requested to put this code into the `main.tf` before using either buyer or seller module.

```tf
# REQUIRED! Enable Anthos Service Mesh
# IMPORTANT! Create only once
resource "google_gke_hub_feature" "servicemesh" {
  provider = google-beta

  location = "global"
  name     = "servicemesh"

  lifecycle {
    create_before_destroy = true
  }
}
```

#### Troubleshooting
While running the  `terraform apply` command on the root folder, you may potentially get an error like this:
```sh
Error: istio-system/istio-ingressgateway failed to fetch resource from kubernetes: context deadline exceeded
```
In parallel, you might see an `ImagePullBackOff` error in the GCP Console log. Such an error may occur because the ASM is not ready before Istio Ingress gets deployed, which may lead to ingress not being able to pull the required image in time.

Try running `terraform apply` once again. It should fix this error.

