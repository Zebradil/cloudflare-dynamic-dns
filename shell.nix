{ pkgs }:

pkgs.mkShell {
  packages = with pkgs; [
    gnused
    go
    go-task
    gofumpt
    goimports-reviser
    golangci-lint
    goreleaser
    gosec
    nix-update
    ytt
  ];
}
