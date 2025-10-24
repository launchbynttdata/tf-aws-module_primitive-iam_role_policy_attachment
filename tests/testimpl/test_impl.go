package testimpl

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposableComplete(t *testing.T, ctx types.TestContext) {
	iamClient := GetAWSIAMClient(t)

	roleName := terraform.Output(t, ctx.TerratestTerraformOptions(), "role_name")
	policyArn := terraform.Output(t, ctx.TerratestTerraformOptions(), "policy_arn")

	t.Run("TestIAMRoleExists", func(t *testing.T) {
		role, err := iamClient.GetRole(context.TODO(), &iam.GetRoleInput{
			RoleName: &roleName,
		})
		require.NoError(t, err, "Failed to get IAM role")
		assert.Equal(t, roleName, *role.Role.RoleName, "Expected role name did not match actual name!")
	})

	t.Run("TestIAMPolicyExists", func(t *testing.T) {
		policy, err := iamClient.GetPolicy(context.TODO(), &iam.GetPolicyInput{
			PolicyArn: &policyArn,
		})
		require.NoError(t, err, "Failed to get IAM policy")
		assert.Equal(t, policyArn, *policy.Policy.Arn, "Expected policy ARN did not match actual ARN!")
	})

	t.Run("TestIAMRolePolicyAttachment", func(t *testing.T) {
		// List policies attached to the role
		attachedPolicies, err := iamClient.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
			RoleName: &roleName,
		})
		require.NoError(t, err, "Failed to list attached role policies")

		// Check if the policy is attached to the role
		policyAttached := false
		for _, attachedPolicy := range attachedPolicies.AttachedPolicies {
			if *attachedPolicy.PolicyArn == policyArn {
				policyAttached = true
				break
			}
		}
		assert.True(t, policyAttached, "Policy %s should be attached to role %s", policyArn, roleName)
	})

	t.Run("TestIAMPolicyAttachmentDetails", func(t *testing.T) {
		// Get policy version to verify policy content
		policy, err := iamClient.GetPolicy(context.TODO(), &iam.GetPolicyInput{
			PolicyArn: &policyArn,
		})
		require.NoError(t, err, "Failed to get IAM policy")

		policyVersion, err := iamClient.GetPolicyVersion(context.TODO(), &iam.GetPolicyVersionInput{
			PolicyArn: &policyArn,
			VersionId: policy.Policy.DefaultVersionId,
		})
		require.NoError(t, err, "Failed to get IAM policy version")

		// AWS returns URL-encoded policy documents, so we need to decode them
		decodedDocument, err := url.QueryUnescape(*policyVersion.PolicyVersion.Document)
		require.NoError(t, err, "Failed to URL decode policy document")

		var policyDoc map[string]interface{}
		err = json.Unmarshal([]byte(decodedDocument), &policyDoc)
		require.NoError(t, err, "Failed to parse policy document")

		// Verify the policy has the expected structure
		assert.Contains(t, policyDoc, "Version", "Policy should have Version field")
		assert.Contains(t, policyDoc, "Statement", "Policy should have Statement field")

		statements, ok := policyDoc["Statement"].([]interface{})
		require.True(t, ok, "Policy should contain Statement array")
		require.Greater(t, len(statements), 0, "Policy should have at least one statement")

		// Verify the first statement contains expected EC2 permissions
		firstStatement := statements[0].(map[string]interface{})
		assert.Contains(t, firstStatement, "Effect", "Statement should have Effect")
		assert.Contains(t, firstStatement, "Action", "Statement should have Action")
		assert.Contains(t, firstStatement, "Resource", "Statement should have Resource")
		assert.Equal(t, "Allow", firstStatement["Effect"], "Effect should be Allow")
	})
}

func GetAWSIAMClient(t *testing.T) *iam.Client {
	awsIAMClient := iam.NewFromConfig(GetAWSConfig(t))
	return awsIAMClient
}

func GetAWSConfig(t *testing.T) (cfg aws.Config) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	require.NoErrorf(t, err, "unable to load SDK config, %v", err)
	return cfg
}
