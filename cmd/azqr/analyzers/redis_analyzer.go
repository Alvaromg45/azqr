package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

type RedisAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
}

func NewRedisAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *RedisAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := RedisAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c RedisAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	redis, err := c.listRedis(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, redis := range redis {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*redis.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *redis.Name,
			Sku:                string(*redis.Properties.SKU.Name),
			Sla:                "TODO",
			Type:               *redis.Type,
			AvailabilityZones:  len(redis.Zones) > 0,
			PrivateEndpoints:   len(redis.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*redis.Name, "kv"),
		})
	}
	return results, nil
}

func (c RedisAnalyzer) listRedis(resourceGroupName string) ([]*armredis.ResourceInfo, error) {
	redisClient, err := armredis.NewClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := redisClient.NewListByResourceGroupPager(resourceGroupName, nil)

	redis := make([]*armredis.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		redis = append(redis, resp.Value...)
	}
	return redis, nil
}
