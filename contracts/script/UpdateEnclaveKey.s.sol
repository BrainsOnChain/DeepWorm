// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {Script} from "forge-std/Script.sol";
import {Worm} from "../src/Worm.sol";

contract UpdateEnclaveKeyScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address wormAddress = vm.envAddress("WORM_ADDRESS");

        bytes memory enclaveKey = vm.envBytes("ENCLAVE_KEY");
        bytes memory seal = vm.envBytes("SEAL");
        uint64 timestampInMilliseconds = uint64(vm.envUint("TIMESTAMP"));

        vm.startBroadcast(deployerPrivateKey);

        Worm worm = Worm(wormAddress);
        worm.updateEnclaveKey(enclaveKey, seal, timestampInMilliseconds);

        vm.stopBroadcast();
    }
}
