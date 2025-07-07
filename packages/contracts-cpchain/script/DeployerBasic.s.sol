// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "./utils/ExistingDeploymentParser.sol";

/**
 * @notice Script used for the first deployment of CpChainLayer core contracts to Cp Chain
 * forge script script/DeployerBasic.s.sol --rpc-url http://127.0.0.1:8545 --private-key $PRIVATE_KEY --broadcast -vvvv
 * forge script script/DeployerBasic.s.sol --rpc-url $RPC_MANTA --private-key $PRIVATE_KEY --broadcast -vvvv
 */
contract DeployerBasic is ExistingDeploymentParser {
    function run() external virtual {
        _parseInitialDeploymentParams("script/configs/Deployment.config.json");

        // START RECORDING TRANSACTIONS FOR DEPLOYMENT
        vm.startBroadcast();

        emit log_named_address("Deployer Address", msg.sender);

        _deployFromScratch();

        // STOP RECORDING TRANSACTIONS FOR DEPLOYMENT
        vm.stopBroadcast();

        // Sanity Checks
        _verifyContractPointers();
        _verifyImplementations();
        _verifyContractsInitialized({isInitialDeployment: true});
        _verifyInitializationParams();

        logAndOutputContractAddresses("script/output/DeploymentBasic.config.json");
    }

    /**
     * @notice Deploy CpChainLayer contracts from scratch for Cp Chain
     */
    function _deployFromScratch() internal {
        // Deploy ProxyAdmin, later set admins for all proxies to be executorMultisig
        cpChainLayerProxyAdmin = new ProxyAdmin(executorMultisig);

        // Set multisigs as pausers, executorMultisig as unpauser
        address[] memory pausers = new address[](3);
        pausers[0] = executorMultisig;
        pausers[1] = operationsMultisig;
        pausers[2] = pauserMultisig;
        address unpauser = executorMultisig;
        cpChainLayerPauserReg = new PauserRegistry(pausers, unpauser);

        /**
         * Deploy upgradeable proxy contracts that **will point** to the implementations. Since the implementation contracts are
         * not yet deployed, we give these proxies an empty contract as the initial implementation, to act as if they have no code.
         */
        emptyContract = new EmptyContract();
        TransparentUpgradeableProxy delegationManagerProxyInstance = new TransparentUpgradeableProxy(address(emptyContract), executorMultisig, "");
        delegationManager = DelegationManager(address(delegationManagerProxyInstance));
        delegationManagerProxyAdmin = ProxyAdmin(getProxyAdminAddress(address(delegationManagerProxyInstance)));
        TransparentUpgradeableProxy cpChainDepositManagerProxyInstance = new TransparentUpgradeableProxy(address(emptyContract), executorMultisig, "");
        cpChainDepositManager = CpChainDepositManager(address(cpChainDepositManagerProxyInstance));
        cpChainDepositManagerProxyAdmin = ProxyAdmin(getProxyAdminAddress(address(cpChainDepositManagerProxyInstance)));
        delegationManagerImplementation = new DelegationManager(cpChainDepositManager);
        cpChainDepositManagerImplementation = new CpChainDepositManager(delegationManager);

        // Upgrade the proxy contracts to point to the implementations
        ICpChainBase[] memory initializeChainBasesToSetDelayBlocks = new ICpChainBase[](0);
        uint256[] memory initializeWithdrawalDelayBlocks = new uint256[](0);

        // DelegationManager
        delegationManagerProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(payable(address(delegationManager))),
            address(delegationManagerImplementation),
            abi.encodeWithSelector(
                DelegationManager.initialize.selector,
                executorMultisig, // initialOwner
                cpChainLayerPauserReg,
                DELEGATION_MANAGER_INIT_PAUSED_STATUS,
                DELEGATION_MANAGER_MIN_WITHDRAWAL_DELAY_BLOCKS,
                initializeChainBasesToSetDelayBlocks,
                initializeWithdrawalDelayBlocks
            )
        );
        // CpChainDepositManager
        cpChainDepositManagerProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(payable(address(cpChainDepositManager))),
            address(cpChainDepositManagerImplementation),
            abi.encodeWithSelector(
                CpChainDepositManager.initialize.selector,
                msg.sender, //initialOwner, set to executorMultisig later after whitelisting strategies
                msg.sender, //initial whitelister, set to STRATEGY_MANAGER_WHITELISTER later
                cpChainLayerPauserReg,
                STRATEGY_MANAGER_INIT_PAUSED_STATUS
            )
        );

        // Deploy ChainBases
        cpChainBaseImplementation = new CpChainBase(cpChainDepositManager);
        // whitelist params
        ICpChainBase[] memory strategiesToWhitelist = new ICpChainBase[](numChainBasesToDeploy);
        bool[] memory thirdPartyTransfersForbiddenValues = new bool[](numChainBasesToDeploy);

        for (uint256 i = 0; i < numChainBasesToDeploy; i++) {
            CpTokenConfig memory chainBaseConfig = strategiesToDeploy[i];

            // Deploy and upgrade chainBase
            TransparentUpgradeableProxy chainBaseBaseProxyInstance = new TransparentUpgradeableProxy(address(emptyContract), executorMultisig, "");
            CpChainBase chainBase = CpChainBase(address(chainBaseBaseProxyInstance));
            chainBaseBaseProxyAdmin = ProxyAdmin(getProxyAdminAddress(address(chainBaseBaseProxyInstance)));
            chainBaseBaseProxyAdmin.upgradeAndCall(
                ITransparentUpgradeableProxy(payable(address(chainBase))),
                address(cpChainBaseImplementation),
                abi.encodeWithSelector(
                    CpChainBase.initialize.selector,
                    IERC20(chainBaseConfig.tokenAddress),
                    cpChainLayerPauserReg,
                    STRATEGY_MAX_PER_DEPOSIT,
                    STRATEGY_MAX_TOTAL_DEPOSITS
                )
            );

            strategiesToWhitelist[i] = chainBase;
            thirdPartyTransfersForbiddenValues[i] = false;

            deployedStrategyArray.push(chainBase);
        }

        // Deploy RewardManager proxy and implementation
        rewardManagerImplementation = new RewardManager(
            delegationManager,
            cpChainDepositManager,
            IERC20(REWARD_MANAGER_RWARD_TOKEN_ADDRESS)
        );
        rewardManager = RewardManager(
            address(
                new TransparentUpgradeableProxy(
                    address(rewardManagerImplementation),
                    address(cpChainLayerProxyAdmin),
                    abi.encodeWithSelector(
                        RewardManager.initialize.selector,
                        executorMultisig,
                        executorMultisig,
                        executorMultisig,
                        REWARD_MANAGER_STAKE_PERCENTAGE
                    )
                )
            )
        );

        // Add strategies to whitelist and set whitelister to STRATEGY_MANAGER_WHITELISTER
        cpChainDepositManager.addChainBasesToDepositWhitelist(strategiesToWhitelist, thirdPartyTransfersForbiddenValues);
        cpChainDepositManager.setStrategyWhitelister(STRATEGY_MANAGER_WHITELISTER);

        // Transfer ownership
        cpChainDepositManager.transferOwnership(executorMultisig);
        cpChainLayerProxyAdmin.transferOwnership(executorMultisig);
    }
}
