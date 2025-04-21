class MakeTray < Formula
    desc "tiny macOS menu‑bar app that lets you launch any Makefile target with a single click"
    homepage "https://github.com/alexrett/make-tray"
    url "https://github.com/alexrett/make-tray/releases/download/v0.0.5/MakeTray_universal"
    sha256 "a4fe50fd77d2f431101725a68612022856d23a7f8d998ec90992afb71494ded0"
    version "0.0.5"
  
    def install
      bin.install "MakeTray_universal" => "make-tray"
    end
  
    test do
      system "#{bin}/make-tray", "--help"
    end
  end