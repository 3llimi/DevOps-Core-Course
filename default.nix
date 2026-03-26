{ pkgs ? import <nixpkgs> {} }:

let
  pythonEnv = pkgs.python3.withPackages (ps: with ps; [
    fastapi
    uvicorn
    pydantic
    starlette
    python-dotenv
    prometheus-client
    annotated-doc
  ]);
in
pkgs.stdenv.mkDerivation rec {
  pname = "devops-info-service";
  version = "1.0.0";
  src = ./.;

  nativeBuildInputs = [ pkgs.makeWrapper ];

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin $out/app
    cp app.py $out/app/app.py

    makeWrapper ${pythonEnv}/bin/python $out/bin/devops-info-service \
      --add-flags "$out/app/app.py"
    runHook postInstall
  '';
}
