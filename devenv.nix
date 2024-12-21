{
  pkgs,
  lib,
  config,
  inputs,
  ...
}: let
  unstable-pkgs = import inputs.unstable-nixpkgs {
    inherit (pkgs.stdenv) system;
  };
in {
  name = "api.pegwings.com";

  languages = {
    go.enable = true;
    nix.enable = true;
  };

  packages = with pkgs; [
    git
    podman
    zsh
    revive
    unstable-pkgs.iferr
    go
    gopls
    impl
    golangci-lint-langserver
    golangci-lint
    templ
    gomodifytags
    gotests
    gotools
    gomarkdoc
    templ
    sqlc
    flyctl
  ];

  scripts = {
    generate.exec = ''
      go generate -v ./...
    '';
    run.exec = ''
      go run main.go
    '';
    dx.exec = ''
      $EDITOR $(git rev-parse --show-toplevel)/devenv.nix
    '';
    deploy.exec = ''
      flyctl deploy
    '';
  };

  enterShell = ''
    git status
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';

  cachix.enable = true;
}
