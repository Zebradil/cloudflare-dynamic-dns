{
  description = "Dynamic DNS client for Cloudflare";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        package = import ./nix/package.nix {
          inherit pkgs self;
        };
      in
      {
        packages.default = package;
        packages.cloudflare-dynamic-dns = package;

        devShells.default = import ./nix/shell.nix { inherit pkgs; };

        # Eval-only sanity check: the NixOS module evaluates with a sample
        # config. NixOS eval only makes sense on Linux systems.
        checks = nixpkgs.lib.optionalAttrs pkgs.stdenv.isLinux {
          nixos-module =
            let
              sys = nixpkgs.lib.nixosSystem {
                inherit system;
                modules = [
                  self.nixosModules.default
                  {
                    # Minimal stubs so config.assertions evaluates (bootloader
                    # and root fs are asserted by NixOS, unrelated to us).
                    fileSystems."/" = { device = "nodev"; fsType = "tmpfs"; };
                    boot.loader.grub.enable = false;
                    system.stateVersion = "24.05";

                    services.cloudflare-dynamic-dns = {
                      enable = true;
                      domains = [ "example.com" ];
                      iface = "eth0";
                      tokenFile = "/dev/null";
                    };
                  }
                ];
              };
              # Forcing ExecStart exercises the options, render.nix and config
              # block; forcing assertions catches the domains/iface/hostId guards.
              failed = nixpkgs.lib.filter (a: !a.assertion) sys.config.assertions;
              execStart = sys.config.systemd.services.cloudflare-dynamic-dns.serviceConfig.ExecStart;
            in
            if failed != [ ] then
              throw "cfddns module assertions failed:\n${nixpkgs.lib.concatMapStringsSep "\n" (a: a.message) failed}"
            else
              pkgs.runCommand "cfddns-nixos-module-eval" { inherit execStart; } "touch $out";
        };
      }
    )
    // {
      nixosModules.default = import ./nix/module/nixos.nix self;
      nixosModules.cloudflare-dynamic-dns = self.nixosModules.default;

      homeManagerModules.default = import ./nix/module/home-manager.nix self;
      homeManagerModules.cloudflare-dynamic-dns = self.homeManagerModules.default;
    };
}
