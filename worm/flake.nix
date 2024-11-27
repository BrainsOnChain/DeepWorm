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
    oyster = {
      url = "github:marlinprotocol/oyster-monorepo";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.fenix.follows = "fenix";
      inputs.naersk.follows = "naersk";
      inputs.nitro-util.follows = "nitro-util";
    };
  };
  outputs = {
    self,
    nixpkgs,
    fenix,
    naersk,
    nitro-util,
    oyster,
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
    worm = naersk'.buildPackage {
      src = ./.;
      nativeBuildInputs = [nematoduino];
    };
    nitro = nitro-util.lib.${system};
    eifArch = "x86_64";
    supervisord = oyster.packages.${system}.musl.external.supervisord.compressed;
    supervisord' = "${supervisord}/bin/supervisord";
    dnsproxy = oyster.packages.${system}.musl.external.dnsproxy.compressed;
    dnsproxy' = "${dnsproxy}/bin/dnsproxy";
    keygen = oyster.packages.${system}.musl.initialization.keygen.compressed;
    keygenEd25519 = "${keygen}/bin/keygen-ed25519";
    tcp-proxy = oyster.packages.${system}.musl.networking.tcp-proxy.compressed;
    itvtProxy = "${tcp-proxy}/bin/ip-to-vsock-transparent";
    vtiProxy = "${tcp-proxy}/bin/vsock-to-ip";
    attestation-server = oyster.packages.${system}.musl.attestation.server.compressed;
    attestationServer = "${attestation-server}/bin/oyster-attestation-server";
    kernels = oyster.packages.${system}.musl.kernels.vanilla;
    kernel = kernels.kernel;
    kernelConfig = kernels.kernelConfig;
    nsmKo = kernels.nsmKo;
    init = kernels.init;
    setup = ./. + "/setup.sh";
    supervisorConf = ./. + "/supervisord.conf";
    app = pkgs.runCommand "app" {} ''
      echo Preparing the app folder
      pwd
      mkdir -p $out
      mkdir -p $out/app
      mkdir -p $out/etc
      cp ${supervisord'} $out/app/supervisord
      cp ${keygenEd25519} $out/app/keygen-ed25519
      cp ${itvtProxy} $out/app/ip-to-vsock-transparent
      cp ${vtiProxy} $out/app/vsock-to-ip
      cp ${attestationServer} $out/app/attestation-server
      cp ${dnsproxy'} $out/app/dnsproxy
      cp ${setup} $out/app/setup.sh
      cp ${worm}/bin/worm $out/app/worm
      chmod +x $out/app/*
      cp ${supervisorConf} $out/etc/supervisord.conf
    '';
    # kinda hacky, my nix-fu is not great, figure out a better way
    initPerms = pkgs.runCommand "initPerms" {} ''
      cp ${init} $out
      chmod +x $out
    '';
  in {
    formatter = {
      "x86_64-linux" = nixpkgs.legacyPackages."x86_64-linux".alejandra;
      "aarch64-linux" = nixpkgs.legacyPackages."aarch64-linux".alejandra;
    };
    packages.${system}.default = nitro.buildEif {
      name = "enclave";
      arch = eifArch;

      init = initPerms;
      kernel = kernel;
      kernelConfig = kernelConfig;
      nsmKo = nsmKo;
      cmdline = builtins.readFile nitro.blobs.${eifArch}.cmdLine;

      entrypoint = "/app/setup.sh";
      env = "";
      copyToRoot = pkgs.buildEnv {
        name = "image-root";
        paths = [app pkgs.busybox pkgs.nettools pkgs.iproute2 pkgs.iptables-legacy pkgs.iperf3];
        pathsToLink = ["/bin" "/app" "/etc"];
      };
    };
  };
}
