class MakeTray < Formula
    desc "tiny macOS menu‑bar app that lets you launch any Makefile target with a single click"
    homepage "https://github.com/alexrett/make-tray"
    url "https://github.com/alexrett/make-tray/releases/download/v0.0.5/MakeTray_universal"
    sha256 "3d3e8712bc1611848cb5244e30c962a5a67aa960cda3d295b3e121e7956cd6cf"
    version "0.0.5"
  
    def install
      bin.install "MakeTray_universal" => "make-tray"
    end
  
    test do
      system "#{bin}/make-tray", "--help"
    end
  end