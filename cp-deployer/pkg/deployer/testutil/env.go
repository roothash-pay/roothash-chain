package testutil

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"runtime"
	"testing"

	"github.com/cpchain-network/cp-chain/common/foundry"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/artifacts"
	op_service "github.com/cpchain-network/cp-chain/cp-service"
	"github.com/cpchain-network/cp-chain/cp-service/testutils"
	"github.com/stretchr/testify/require"
)

func LocalArtifacts(t *testing.T) (*artifacts.Locator, foundry.StatDirFs) {
	_, testFilename, _, ok := runtime.Caller(0)
	require.Truef(t, ok, "failed to get test filename")
	monorepoDir, err := op_service.FindMonorepoRoot(testFilename)
	require.NoError(t, err)
	artifactsDir := path.Join(monorepoDir, "packages", "contracts-bedrock", "forge-artifacts")
	artifactsURL, err := url.Parse(fmt.Sprintf("file://%s", artifactsDir))
	require.NoError(t, err)
	loc := &artifacts.Locator{
		URL: artifactsURL,
	}

	testCacheDir := testutils.IsolatedTestDirWithAutoCleanup(t)

	artifactsFS, err := artifacts.Download(context.Background(), loc, artifacts.NoopProgressor(), testCacheDir)
	require.NoError(t, err)

	return loc, artifactsFS
}
