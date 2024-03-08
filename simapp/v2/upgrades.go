package simapp

import (
	"context"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/x/accounts"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v0.50.x to v0.51.x
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v0.51.x.
const UpgradeName = "v050-to-v051"

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (appmodulev2.VersionMap, error) {
			return app.ModuleManager().RunMigrations(ctx, fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{
			Add: []string{
				accounts.ModuleName,
				protocolpooltypes.ModuleName,
			},
			Delete: []string{
				crisistypes.ModuleName, // The SDK discontinued the crisis module in v0.51.0
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		_ = storeUpgrades
		// app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}