{
  description = "A golang cli git commit generator.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?tag=24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages = {
          gita = pkgs.buildGoModule {
            pname = "gita";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-TnCEtzKZLE1GrqWvmSTkTYhOJumbjAabrZA3EOg7my4=";
          };
          default = self.packages.${system}.gita;
        };
        defaultPackage = self.packages.${system}.gita;
      }
    );
}
