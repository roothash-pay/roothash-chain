// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Vm.sol";
import { console, Script } from "forge-std/Script.sol";


import { WCP } from "../src/core/token/WCP.sol";



contract DeployerWrappedCpScript is Script {
    WCP public wCp;

    function run() public {
        vm.startBroadcast();
        wCp = new WCP();
        console.log("deploy wCp:", address(wCp));
        console.log("wCp", wCp.name());
        vm.stopBroadcast();
    }
}
