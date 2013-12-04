package utils

import "testing"

func Test_buildPath_1( t *testing.T) {
    if buildPath("/etc", "pwd") != "/etc/pwd" {
        t.Error("buildPath didn't work as expected")
    } else {
        t.Log("one test passed")
    }
}

func Test_determineFileType_1(t *testing.T) {
    fn := determineFileType()
    if fn("casablanca.avi") != "video" {
        t.Error("casablanca.avi should be video")
    }
}

func Test_determineFileType_2(t *testing.T) {
    fn := determineFileType()
    if fn("Mulholland.Dr.2001.720p.x264.AAC.300mbunited.com-scOrp.mkv") != "video" {
      t.Error("Should be video")
    }
}

func Test_determineFileType_3(t *testing.T) {
    fn := determineFileType()
    if fn("The.Maltese.Falcon.1941.512x368.25fps.922kbs.88mp3.MultiSub.WunSeeDee.avi") != "video" {
        t.Error("Should be video")
    }
}

func Test_determineFileType_4(t *testing.T) {
    fn := determineFileType()
    if fn("01 - (Fuel Of My Soul).mp3") != "audio" {
        t.Error("should be audio")
    }
}

func Test_determineFileName_1(t *testing.T) {
    fn := determineFileName()
    if fn("folder/01 - (Fule of My Sould).mp3") != "01 - (Fule of My Sould)" {
        t.Error("bad filename returned")
    }
    t.Log(fn("01 - (Fuel Of My Soul).mp3"))
}

func Test_determineFileName_2(t *testing.T) {
    fn := determineFileName()
    if fn("/The.Maltese.Falcon.1941.512x368.25fps.922kbs.88mp3.MultiSub.WunSeeDee.avi") != "The.Maltese.Falcon.1941.512x368.25fps.922kbs.88mp3.MultiSub.WunSeeDee" {
        t.Error("bad filename returned")
    }
    t.Log(fn("The.Maltese.Falcon.1941.512x368.25fps.922kbs.88mp3.MultiSub.WunSeeDee.avi"))
}

