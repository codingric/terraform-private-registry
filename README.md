# terraform-private-registry

This application provides a Terraform Registry compatible API to serve terraform providers. It designed with a S3 bucket backend, it provides download links via presigned S3 links.

## Build and Run locally

```shell
docker build --platform linux/amd64 -t terraform-private-registry .
docker run -p "8080:8080" -e BUCKET_NAME -v "$HOME/.aws :/root/.aws" terraform-private-registry
```

### API Endpoints

All expected endpoints follow the terraform provider registry protocol [See documentaion](https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#service-discovery). In addtion to this there is also an endpoint to upload a binary via POST.

```shell
curl -F provider=@{{provider-binary}} https://{{registry-address}}.com/v1/providers/{{namespace}}/{{name}}/{{version}}/upload/{{os}}/{{arch}}
```

Uploading will do the following:
- Zip the binary and rename with {{provier-name}}\_{{version}}_{{os}}\_{{arch}}.zip
- Generate SHA256 sum file
- GPG Sign the sum file
- Update provider information with new versions