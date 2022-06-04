{
  inputs.nixpkgs.url = "github:nixOS/nixpkgs/nixos-24.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }@inputs:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      rec {
        packages.kjudge-templates = (pkgs.callPackage ./nix/frontend.nix { });
        packages.kjudge = (
          pkgs.callPackage ./nix/backend.nix {
            templates = packages.kjudge-templates;
          }
        );
        devShell = pkgs.mkShell {
          inputsFrom = [
            packages.kjudge
            packages.kjudge-templates
          ];
        };
      }
    );
}
