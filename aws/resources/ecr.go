package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/gruntwork-io/cloud-nuke/config"
	"github.com/gruntwork-io/cloud-nuke/logging"
	"github.com/gruntwork-io/cloud-nuke/report"
	"github.com/gruntwork-io/go-commons/errors"
)

func (registry *ECR) getAll(c context.Context, configObj config.Config) ([]*string, error) {
	var repositoryNames []*string

	paginator := ecr.NewDescribeRepositoriesPaginator(registry.Client, &ecr.DescribeRepositoriesInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(c)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}

		for _, repository := range output.Repositories {
			if configObj.ECRRepository.ShouldInclude(config.ResourceValue{
				Time: repository.CreatedAt,
				Name: repository.RepositoryName,
			}) {
				repositoryNames = append(repositoryNames, repository.RepositoryName)
			}
		}
	}

	return repositoryNames, nil
}

func (registry *ECR) nukeAll(repositoryNames []string) error {
	if len(repositoryNames) == 0 {
		logging.Debugf("No ECR repositories to nuke in region %s", registry.Region)
		return nil
	}

	var deletedNames []*string

	for _, repositoryName := range repositoryNames {
		params := &ecr.DeleteRepositoryInput{
			Force:          true,
			RepositoryName: aws.String(repositoryName),
		}

		_, err := registry.Client.DeleteRepository(registry.Context, params)

		// Record status of this resource
		e := report.Entry{
			Identifier:   repositoryName,
			ResourceType: "ECR Repository",
			Error:        err,
		}
		report.Record(e)

		if err != nil {
			logging.Debugf("[Failed] %s", err)
		} else {

			deletedNames = append(deletedNames, aws.String(repositoryName))
			logging.Debugf("Deleted ECR Repository: %s", repositoryName)
		}
	}

	logging.Debugf("[OK] %d ECR Repositories deleted in %s", len(deletedNames), registry.Region)

	return nil
}
