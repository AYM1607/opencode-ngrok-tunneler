{
  description = "OpenCode Tunneler flake";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
  };

  outputs =
    { nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        opencode-tunneler = pkgs.callPackage ./opencode-tunneler.nix { };
      in
      {
        packages.default = opencode-tunneler;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gotools
            gopls
          ];
        };
      }
    );
}
