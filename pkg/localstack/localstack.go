/*
This package was written to help writing tests with Localstack.  
(https://github.com/localstack/localstack)  It uses libraries that help create
and manage a Localstack docker container for your go tests.

Requirements

    Go v1.11.0 or higher 
    Docker (Tested on version 19.03.0-rc Community Edition)
*/
package localstack

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"strings"
)

// Localstack_Repository is the Localstack Docker repository
const Localstack_Repository string = "localstack/localstack"
// Localstack_Tag is the last tested version of the Localstack Docker repository
const Localstack_Tag string = "0.9.1"

// Localstack is a structure used to control the lifecycle of the Localstack 
// Docker container.
type Localstack struct {
    // Resource is a pointer to the dockertest.Resource 
    // object that is the localstack docker container.
    // (https://godoc.org/github.com/ory/dockertest#Resource)
	Resource *dockertest.Resource
    // Services is a pointer to a collection of service definitions
    // that are being requested from this particular instance of Localstack.
	Services *LocalstackServiceCollection
}

// Destroy simply shuts down and cleans up the Localstack container out of docker.
func (ls *Localstack) Destroy() error {
	
	pool, err := dockertest.NewPool("")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(ls.Resource); err != nil {
		return errors.New(fmt.Sprintf("Could not purge resource: %s", err))
	}

	return nil
}

// EndpointResolver is necessary to route traffic to AWS services in your code to the Localstack
// endpoints.
func (l *Localstack) EndpointResolver() aws.EndpointResolverFunc {
	return aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == dynamodb.ServiceID &&
			l.Services.Contains("dynamodb") {
			return aws.Endpoint{
				URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4569/tcp")),
			}, nil
		} else if service == dynamodbstreams.ServiceID &&
			l.Services.Contains("dynamodbstreams") {
			return aws.Endpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4570/tcp")) }, nil
		} else if service == apigateway.ServiceID &&
			l.Services.Contains("apigateway") {
			return aws.Endpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4567/tcp")) }, nil
		} else if service == kinesis.ServiceID &&
			l.Services.Contains("kinesis") {
			return aws.Endpoint{URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4568/tcp"))}, nil
		} else if service == s3.ServiceID &&
			l.Services.Contains("s3") {
			return aws.Endpoint{URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4572/tcp"))}, nil
		}
		//} else if service == endpoints.EsServiceID &&
		//	l.Services.Contains("es")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4571/tcp")) }, nil
		//} else if service == endpoints.FirehoseServiceID &&
		//	l.Services.Contains("firehose")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4573/tcp")) }, nil
		//} else if service == endpoints.LambdaServiceID &&
		//	l.Services.Contains("lambda")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4574/tcp")) }, nil
		//} else if service == endpoints.SnsServiceID &&
		//	l.Services.Contains("sns")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4575/tcp")) }, nil
		//} else if service == endpoints.SqsServiceID &&
		//	l.Services.Contains("sqs")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4576/tcp")) }, nil
		//} else if service == endpoints.RedshiftServiceID &&
		//	l.Services.Contains("redshift")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4577/tcp")) }, nil
		//} else if service == endpoints.EmailServiceID &&
		//	l.Services.Contains("ses")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4579/tcp")) }, nil
		//} else if service == endpoints.Route53ServiceID &&
		//	l.Services.Contains("route53")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4580/tcp")) }, nil
		//} else if service == endpoints.CloudformationServiceID &&
		//	l.Services.Contains("cloudformation")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4581/tcp")) }, nil
		//} else if service == endpoints.MonitoringServiceID &&
		//	l.Services.Contains("cloudwatch")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4582/tcp")) }, nil
		//} else if service == endpoints.SsmServiceID &&
		//	l.Services.Contains("ssm")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4583/tcp")) }, nil
		//} else if service == endpoints.SecretsmanagerServiceID &&
		//	l.Services.Contains("secretsmanager")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4584/tcp")) }, nil
		//} else if service == endpoints.StatesServiceID &&
		//	l.Services.Contains("stepfunctions")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4585/tcp")) }, nil
		//} else if service == endpoints.LogsServiceID &&
		//	l.Services.Contains("logs")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4586/tcp")) }, nil
		//} else if service == endpoints.StsServiceID &&
		//	l.Services.Contains("sts")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4592/tcp")) }, nil
		//} else if service == endpoints.IamServiceID &&
		//	l.Services.Contains("iam")  {
		//	return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4593/tcp")) }, nil
		//} else {
		//	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
		//}


		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
}

func (l *Localstack) CreateConfig() (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("a", "b", "c")),
		config.WithEndpointResolver(l.EndpointResolver()))
}

// NewLocalstack creates a new Localstack docker container based on the latest version.
func NewLocalstack(services *LocalstackServiceCollection) (*Localstack, error) {
	return NewSpecificLocalstack(services, "", Localstack_Repository, "latest")
}

// NewSpecificLocalstack creates a new Localstack docker container based on
// the given name, repository, and tag given.  NOTE:  The Docker image used should be a 
// Localstack image.  The behavior is unknown otherwise.  This method is provided
// to allow special situations like using a tag other than latest or when referencing 
// an internal Localstack image.
func NewSpecificLocalstack(services *LocalstackServiceCollection, name, repository, tag string) (*Localstack, error) {
	return newLocalstack(services, &_DockerWrapper{ }, name, repository, tag)
}

func getLocalstack(services *LocalstackServiceCollection, dockerWrapper DockerWrapper, name, repository, tag string) (*dockertest.Resource, error) {

    if name != "" {
        containers, err := dockerWrapper.ListContainers(docker.ListContainersOptions { All: true })
        if err != nil {
            return nil, errors.New(fmt.Sprintf("Unable to retrieve docker containers: %s", err))
        }
        for _, c := range containers {
            if c.Image == fmt.Sprintf("%s:%s", repository, tag) {
                for _,internalName := range c.Names {
                    if internalName == fmt.Sprintf("/%s", name) {
                        container, err := dockerWrapper.InspectContainer(c.ID)
                        if err !=  nil {
                            return nil, errors.New(fmt.Sprintf("Unable to inspect container %s: %s", c.ID, err))
                        }
                        return &dockertest.Resource{ Container: container }, nil
                    }
                }
            }
        }
    }

	return nil, nil
}

func newLocalstack(services *LocalstackServiceCollection, wrapper DockerWrapper, name, repository, tag string) (*Localstack, error) {

	localstack, err := getLocalstack(services, wrapper, name, repository, tag)
	if err != nil {
		return nil, err	
	}

	if localstack == nil {

		// Fifth, If we didn't find a running container before, we spin one up now.
		localstack, err = wrapper.RunWithOptions(&dockertest.RunOptions{
			Repository: repository,
			Tag: tag,
            Name: name, //If name == "", docker ignores it.
			Env: []string{
				fmt.Sprintf("SERVICES=%s", services.GetServiceMap()),
			},
		})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not start resource: %s", err))
		}
	}

	// Sixth, we wait for the services to be ready before we allow the tests
	// to be run.
	for _, service := range *services {
		if err := wrapper.Retry(func() error {

			// We have to use a method that checks the output
			// of the docker container here because simply checking for
			// connetivity on the ports doesn't work.
			client, err := docker.NewClientFromEnv()
			if err != nil {
				return errors.New(fmt.Sprintf("Unable to create a docker client: %s", err))
			}

			buffer := new(bytes.Buffer)

			logsOptions := docker.LogsOptions {
				Container: localstack.Container.ID,
				OutputStream: buffer,
				RawTerminal: true,
				Stdout: true,
				Stderr: true,
			}
			err = client.Logs(logsOptions)
			if err != nil {
				return errors.New(fmt.Sprintf("Unable to retrieve logs for container %s: %s", localstack.Container.ID, err))
			}

			scanner := bufio.NewScanner(buffer)
			for scanner.Scan() {
				token := strings.TrimSpace(scanner.Text())
				expected := "Ready."
				if strings.Contains(strings.TrimSpace(token),expected) {
					return nil
				}
			}
			if err := scanner.Err(); err != nil {
				return errors.New(fmt.Sprintf("Reading input: %s", err))
			}
			return errors.New("Not Ready")
		}); err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to connect to %s: %s", service.Name, err))
		}
	}

	return &Localstack{
		Resource: localstack,
		Services: services,
	}, nil
}

