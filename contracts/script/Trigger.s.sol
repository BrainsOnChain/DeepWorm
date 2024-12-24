// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Script} from "forge-std/Script.sol";
import {Worm} from "../src/Worm.sol";

contract UpdateEnclaveKeyScript is Script {
    function run() external {
        uint256 callerPrivateKey = vm.envUint("PRIVATE_KEY");
        address wormAddress = vm.envAddress("WORM_ADDRESS");

        vm.startBroadcast(callerPrivateKey);

        Worm worm = Worm(wormAddress);
        worm.trigger();

        vm.stopBroadcast();
    }
}
