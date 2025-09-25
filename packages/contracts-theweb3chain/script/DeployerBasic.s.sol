// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "./utils/ExistingDeploymentParser.sol";

/**
 * @notice Script used for the first deployment of theweb3ChainLayer core contracts to theweb3 chain
 * forge script script/DeployerBasic.s.sol --rpc-url http://127.0.0.1:8545 --private-key $PRIVATE_KEY --broadcast -vvvv
 * forge script script/DeployerBasic.s.sol --rpc-url $RPC_URL --private-key $PRIVATE_KEY --broadcast -vvvv
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
//        _verifyContractPointers();
//        _verifyImplementations();
//        _verifyContractsInitialized();
//        _verifyInitializationParams();

        logAndOutputContractAddresses(
            "script/output/DeploymentBasic.config.json"
        );
    }

    /**
     * @notice Deploy theweb3ChainLayer contracts from scratch for theweb3 Chain
     */
    function _deployFromScratch() internal {
        // Deploy ProxyAdmin, later set admins for all proxies to be executorMultisig
        theweb3ChainLayerProxyAdmin = new ProxyAdmin(executorMultisig);

        // Set multisigs as pausers, executorMultisig as unpauser
        address[] memory pausers = new address[](3);
        pausers[0] = executorMultisig;
        pausers[1] = operationsMultisig;
        pausers[2] = pauserMultisig;
        address unpauser = executorMultisig;
        theweb3ChainLayerPauserReg = new PauserRegistry(pausers, unpauser);

        emptyContract = new EmptyContract();

        // Deploy and upgrade chainBase
        TransparentUpgradeableProxy chainBaseBaseProxyInstance = new TransparentUpgradeableProxy(
                address(emptyContract),
                executorMultisig,
                ""
            );
        theweb3ChainBase = theweb3ChainBase(payable(address(chainBaseBaseProxyInstance)));
        chainBaseBaseProxyAdmin = ProxyAdmin(
            getProxyAdminAddress(address(chainBaseBaseProxyInstance))
        );

        TransparentUpgradeableProxy delegationManagerProxyInstance = new TransparentUpgradeableProxy(
                address(emptyContract),
                executorMultisig,
                ""
            );
        delegationManager = DelegationManager(
            payable(address(delegationManagerProxyInstance))
        );
        delegationManagerProxyAdmin = ProxyAdmin(
            getProxyAdminAddress(address(delegationManagerProxyInstance))
        );

        TransparentUpgradeableProxy theweb3ChainDepositManagerProxyInstance = new TransparentUpgradeableProxy(
                address(emptyContract),
                executorMultisig,
                ""
            );
        theweb3ChainDepositManager = theweb3ChainDepositManager(
            payable(address(theweb3ChainDepositManagerProxyInstance))
        );
        theweb3ChainDepositManagerProxyAdmin = ProxyAdmin(
            getProxyAdminAddress(address(theweb3ChainDepositManagerProxyInstance))
        );

        TransparentUpgradeableProxy rewardManagerProxyInstance = new TransparentUpgradeableProxy(
                address(emptyContract),
                executorMultisig,
                ""
            );
        rewardManager = RewardManager(
            payable(address(rewardManagerProxyInstance))
        );
        rewardManagerProxyAdmin = ProxyAdmin(
            getProxyAdminAddress(address(rewardManagerProxyInstance))
        );

        TransparentUpgradeableProxy slashingManagerProxyInstance = new TransparentUpgradeableProxy(
                address(emptyContract),
                executorMultisig,
                ""
            );
        slashingManager = SlashingManager(
            payable(address(slashingManagerProxyInstance))
        );
        slashingManagerProxyAdmin = ProxyAdmin(
            getProxyAdminAddress(address(slashingManagerProxyInstance))
        );

        delegationManagerImplementation = new DelegationManager();
        theweb3ChainDepositManagerImplementation = new theweb3ChainDepositManager();
        theweb3ChainBaseImplementation = new theweb3ChainBase();
        rewardManagerImplementation = new RewardManager();
        slashingManagerImplementation = new SlashingManager();

        // DelegationManager
        delegationManagerProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(payable(address(delegationManager))),
            address(delegationManagerImplementation),
            abi.encodeWithSelector(
                DelegationManager.initialize.selector,
                executorMultisig, // initialOwner
                theweb3ChainLayerPauserReg,
                DELEGATION_MANAGER_INIT_PAUSED_STATUS,
                0,
                theweb3ChainDepositManager,
                theweb3ChainBase,
                slashingManager
            )
        );
        // theweb3ChainDepositManager
        theweb3ChainDepositManagerProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(
                payable(address(theweb3ChainDepositManager))
            ),
            address(theweb3ChainDepositManagerImplementation),
            abi.encodeWithSelector(
                theweb3ChainDepositManager.initialize.selector,
                msg.sender, //initialOwner, set to executorMultisig later after whitelisting strategies
                address(delegationManagerProxyInstance),
                address(chainBaseBaseProxyInstance)
            )
        );

        // theweb3ChainBase
        chainBaseBaseProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(payable(address(theweb3ChainBase))),
            address(theweb3ChainBaseImplementation),
            abi.encodeWithSelector(
                theweb3ChainBase.initialize.selector,
                theweb3ChainLayerPauserReg,
                theweb3ChainBASE_MIN_DEPOSIT,
                theweb3ChainBASE_MAX_DEPOSIT,
                theweb3ChainDepositManager
            )
        );

//        // Deploy RewardManager proxy and implementation
//        rewardManagerProxyAdmin.upgradeAndCall(
//            ITransparentUpgradeableProxy(payable(address(rewardManager))),
//            address(rewardManagerImplementation),
//            abi.encodeWithSelector(
//                RewardManager.initialize.selector,
//                executorMultisig,
//                executorMultisig,
//                executorMultisig,
//                REWARD_MANAGER_STAKE_PERCENTAGE,
//                theweb3ChainLayerPauserReg
//            )
//        );

        slashingManagerProxyAdmin.upgradeAndCall(
            ITransparentUpgradeableProxy(payable(address(slashingManager))),
            address(slashingManagerImplementation),
            abi.encodeWithSelector(
                SlashingManager.initialize.selector,
                executorMultisig,
                delegationManager,
                executorMultisig,
                1000,
                executorMultisig
            )
        );

        // Transfer ownership
        theweb3ChainDepositManager.transferOwnership(executorMultisig);
        theweb3ChainLayerProxyAdmin.transferOwnership(executorMultisig);
    }
}
