let
  # Import nixpkgs if needed
  pkgs = import <nixpkgs> {};
in
  {
    lib ? pkgs.lib,
    buildGoModule ? pkgs.buildGoModule,
    fetchFromGitHub ? pkgs.fetchFromGitHub,
    installShellFiles ? pkgs.installShellFiles,
    # version and vendorSha256 should be specified by the caller
    version ? "latest",
    vendorHash,
  }:
    buildGoModule rec {
      pname = "kami-chan-server-installer";
      inherit version vendorSha256;

      src = ./..;

      nativeBuildInputs = [
        installShellFiles
      ];

      meta = with lib; {
        description = "A command line tool installing Paper servers";
        license = licenses.mit;
        mainProgram = "kami-chan-server-installer";
      };
    }