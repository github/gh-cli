package set

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmd/secret/shared"
)

type SecretPayload struct {
	EncryptedValue string `json:"encrypted_value"`
	Visibility     string `json:"visibility,omitempty"`
	Repositories   []int  `json:"selected_repository_ids,omitempty"`
	KeyID          string `json:"key_id"`
}

func putSecret(client *api.Client, host, path string, payload SecretPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize: %w", err)
	}
	requestBody := bytes.NewReader(payloadBytes)

	return client.REST(host, "PUT", path, requestBody, nil)
}

func putOrgSecret(client *api.Client, host string, pk *shared.PubKey, opts SetOptions, eValue string) error {
	secretName := opts.SecretName
	orgName := opts.OrgName
	visibility := opts.Visibility

	var repositoryIDs []int
	var err error
	if orgName != "" && visibility == shared.Selected {
		repositoryIDs, err = mapRepoNameToID(client, host, orgName, opts.RepositoryNames)
		if err != nil {
			return fmt.Errorf("failed to look up IDs for repositories %v: %w", opts.RepositoryNames, err)
		}
	}

	payload := SecretPayload{
		EncryptedValue: eValue,
		KeyID:          pk.ID,
		Repositories:   repositoryIDs,
		Visibility:     visibility,
	}
	path := fmt.Sprintf("orgs/%s/actions/secrets/%s", orgName, secretName)

	return putSecret(client, host, path, payload)
}

func putRepoSecret(client *api.Client, pk *shared.PubKey, repo ghrepo.Interface, secretName, eValue string) error {
	payload := SecretPayload{
		EncryptedValue: eValue,
		KeyID:          pk.ID,
	}
	path := fmt.Sprintf("repos/%s/actions/secrets/%s", ghrepo.FullName(repo), secretName)
	return putSecret(client, repo.RepoHost(), path, payload)
}

// This does similar logic to `api.RepoNetwork`, but without the overfetching.
func mapRepoNameToID(client *api.Client, host, orgName string, repositoryNames []string) ([]int, error) {
	queries := make([]string, 0, len(repositoryNames))
	for i, repoName := range repositoryNames {
		queries = append(queries, fmt.Sprintf(`
			repo_%03d: repository(owner: %q, name: %q) {
				databaseId
			}
		`, i, orgName, repoName))
	}

	query := fmt.Sprintf(`query MapRepositoryNames { %s }`, strings.Join(queries, ""))

	graphqlResult := make(map[string]*struct {
		DatabaseID int `json:"databaseId"`
	})

	err := client.GraphQL(host, query, nil, &graphqlResult)

	gqlErr, isGqlErr := err.(*api.GraphQLErrorResponse)
	if isGqlErr {
		for _, ge := range gqlErr.Errors {
			if ge.Type == "NOT_FOUND" {
				return nil, fmt.Errorf("could not find %s/%s", orgName, ge.Path[0])
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to look up repositories: %w", err)
	}

	repoKeys := make([]string, 0, len(repositoryNames))
	for k := range graphqlResult {
		repoKeys = append(repoKeys, k)
	}
	sort.Strings(repoKeys)

	result := make([]int, len(repositoryNames))
	for i, k := range repoKeys {
		result[i] = graphqlResult[k].DatabaseID
	}

	return result, nil
}
