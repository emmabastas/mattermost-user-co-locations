{ pkgs ? import <nixpkgs> {} }:

let
  unstablePackages = import (pkgs.fetchFromGitHub {
    owner = "NixOS";
    repo = "nixpkgs";
    #rev = "07bf3d25ce1da3bee6703657e6a787a4c6cdcea9";
    #sha256 = "0v5yfcc8ml58lfwmmca3vrilp6a9wqk2ak56gg1y4idmm9j8f6l3";
    rev = "49a2bcc6e2065909c701f862f9a1a62b3082b40a";
    sha256 = "0v5yfcc8ml58lfwmmca3vrilp6a9wqk2ak56gg1y4idmm9j8f6l3";
  }) {};

  sbcl = unstablePackages.lispPackages_new.sbclWithPackages (ps: [
    ps.alexandria
    ps.jzon
  ]);
in
pkgs.mkShell {

  buildInputs = with pkgs; [
    go
    gopls
    sbcl
  ];
}
