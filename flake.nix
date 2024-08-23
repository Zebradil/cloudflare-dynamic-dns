{
  description = "Dynamic DNS client for Cloudflare";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixpkgs-master.url = "github:NixOS/nixpkgs/master";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = {
    self,
    nixpkgs,
    nixpkgs-master,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {inherit system;};
      pkgsMaster = import nixpkgs-master {inherit system;};
      baseVersion = "4.3.0";
      commit =
        if (self ? shortRev)
        then self.shortRev
        else if (self ? dirtyShortRev)
        then self.dirtyShortRev
        else "unknown";
      version = "${baseVersion}-${commit}";
      defaultPackage = pkgs.buildGoModule {
        CGO_ENABLED = "0";
        pname = "cloudflare-dynamic-dns";
        src = ./.;
        vendorHash = "sha256-/UaTOCbE8ieCtME6AudbXE5ntCptPFoESYrdn7qK0MU=";
        version = version;

        ldflags = [
          "-s"
          "-w"
          "-X=main.version=${baseVersion}"
          "-X=main.commit=${commit}"
          "-X=main.date=1970-01-01"
        ];

        meta = {
          changelog = "https://github.com/Zebradil/cloudflare-dynamic-dns/blob/${baseVersion}/CHANGELOG.md";
          description = "Dynamic DNS client for Cloudflare";
          homepage = "https://github.com/Zebradil/cloudflare-dynamic-dns";
          license = nixpkgs.lib.licenses.mit;
        };
      };
    in {
      packages.cloudflare-dynamic-dns = defaultPackage;
      defaultPackage = defaultPackage;

      # Provide an application entry point
      apps.default = flake-utils.lib.mkApp {
        drv = defaultPackage;
      };

      devShells.default = pkgs.mkShell {
        packages =
          (with pkgs; [
            # TODO: add semantic-release and plugins
            gnused
            go
            go-task
            gofumpt
            goimports-reviser
            golangci-lint
            gosec
            nix-update
            ytt
          ])
          ++ [
            defaultPackage
            pkgsMaster.goreleaser
          ];
      };
    });
}
