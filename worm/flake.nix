{
  nixConfig = {
    extra-substituters = ["https://nix-cache.marlin.org/oyster"];
    extra-trusted-public-keys = ["oyster:UL7iDKjSdB6YNPArz1JSuca7yJJWPuzz/SXtTgvFr7o="];
  };
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/24.05";
    fenix = {
      url = "github:nix-community/fenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    naersk = {
      url = "github:nix-community/naersk";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nitro-util = {
      url = "github:monzo/aws-nitro-util";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  outputs = {
    self,
    nixpkgs,
    fenix,
    naersk,
    nitro-util,
  }: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
    target = "x86_64-unknown-linux-gnu";
    toolchain = with fenix.packages.${system};
      combine [
        stable.cargo
        stable.rustc
        targets.${target}.stable.rust-std
      ];
    naersk' = naersk.lib.${system}.override {
      cargo = toolchain;
      rustc = toolchain;
    };
    nematoduino = pkgs.stdenv.mkDerivation {
      name = "nematoduino";
      src = ./.;
      nativeBuildInputs = [pkgs.cmake];
      cmakeFlags = [
        "-DCMAKE_BUILD_TYPE=Release"
      ];
    };
  in {
    formatter = {
      "x86_64-linux" = nixpkgs.legacyPackages."x86_64-linux".alejandra;
      "aarch64-linux" = nixpkgs.legacyPackages."aarch64-linux".alejandra;
    };
    worm = naersk'.buildPackage {
      src = ./.;
      nativeBuildInputs = [nematoduino];
    };
  };
}
