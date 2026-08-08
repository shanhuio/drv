package homerelease

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shanhu.io/g/httputil"
)

// fakeObjServer is an httptest server that implements the object "exists" and
// upload (PUT) APIs the Uploader talks to.
type fakeObjServer struct {
	server *httptest.Server

	existing map[string]bool // keys the server already has

	checked  []string          // keys the client asked "exists" for
	uploaded map[string][]byte // keys the client PUT, with their bodies
	putFail  bool              // if set, PUT returns 500
}

func newFakeObjServer(existing map[string]bool) *fakeObjServer {
	s := &fakeObjServer{
		existing: existing,
		uploaded: make(map[string][]byte),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/obj/exists", func(
		w http.ResponseWriter, r *http.Request,
	) {
		var key string
		if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.checked = append(s.checked, key)
		_ = json.NewEncoder(w).Encode(s.existing[key])
	})
	mux.HandleFunc("/obj/", func(
		w http.ResponseWriter, r *http.Request,
	) {
		if r.Method != http.MethodPut {
			http.Error(w, "not put", http.StatusMethodNotAllowed)
			return
		}
		if s.putFail {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/obj/")
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.uploaded[key] = bs
	})

	s.server = httptest.NewServer(mux)
	return s
}

func (s *fakeObjServer) uploader() *Uploader {
	return &Uploader{
		Client:  httputil.NewClientMust(s.server.URL),
		DataURL: "/obj",
		APIURL:  "/api/obj",
	}
}

func (s *fakeObjServer) close() { s.server.Close() }

// makeObjsTar writes each named source file with the given content, keys it by
// the real sha256 checksum of that content (as the build does), and packs the
// files with a manifest into an objs.tar via writeObjects. It returns the
// objs.tar path and a map from file name to its "sha256:<hex>" checksum key.
func makeObjsTar(
	t *testing.T, dir string, files map[string]string,
) (string, map[string]string) {
	t.Helper()

	objects := make(map[string]string)
	sums := make(map[string]string)
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatalf("write src %q: %v", p, err)
		}
		h := sha256.Sum256([]byte(content))
		sum := "sha256:" + hex.EncodeToString(h[:])
		objects[sum] = p
		sums[name] = sum
	}

	out := filepath.Join(dir, "objs.tar")
	if err := writeObjects(out, objects, dir); err != nil {
		t.Fatalf("writeObjects: %v", err)
	}
	return out, sums
}

// checkDigest verifies that key is a well-formed "sha256:<hex>" digest whose
// hash matches the given body.
func checkDigest(t *testing.T, key string, body []byte) {
	t.Helper()
	const prefix = "sha256:"
	hexPart, ok := strings.CutPrefix(key, prefix)
	if !ok {
		t.Errorf("digest %q missing %q prefix", key, prefix)
		return
	}
	if len(hexPart) != sha256.Size*2 {
		t.Errorf("digest %q has %d hex chars, want %d",
			key, len(hexPart), sha256.Size*2)
	}
	h := sha256.Sum256(body)
	want := hex.EncodeToString(h[:])
	if hexPart != want {
		t.Errorf("digest %q does not match content sum sha256:%s", key, want)
	}
}

func TestUploaderUpload(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"core.tar.gz":        "content-a",
		"nextcloud33.tar.gz": "content-b",
	}
	objs, sums := makeObjsTar(t, dir, files)
	coreSum := sums["core.tar.gz"]
	ncSum := sums["nextcloud33.tar.gz"]

	// core already exists and should be skipped.
	srv := newFakeObjServer(map[string]bool{coreSum: true})
	defer srv.close()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	if err := srv.uploader().Upload(objs); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	// The existing object is not uploaded; the missing one is, with its bytes.
	if _, ok := srv.uploaded[coreSum]; ok {
		t.Errorf("existing object %q should not be uploaded", coreSum)
	}
	if got := string(srv.uploaded[ncSum]); got != "content-b" {
		t.Errorf("uploaded %q = %q, want %q", ncSum, got, "content-b")
	}
	if len(srv.uploaded) != 1 {
		t.Errorf("uploaded %d objects, want 1: %v", len(srv.uploaded), srv.uploaded)
	}

	// Every uploaded object is keyed by a valid sha256 digest of its content.
	for key, body := range srv.uploaded {
		checkDigest(t, key, body)
	}

	// The manifest is neither existence-checked nor uploaded.
	if _, ok := srv.uploaded[objectManifestName]; ok {
		t.Errorf("manifest should not be uploaded")
	}
	for _, k := range srv.checked {
		if k == objectManifestName {
			t.Errorf("manifest should not be existence-checked")
		}
	}
	// Both real objects are existence-checked.
	wantChecked := map[string]bool{coreSum: true, ncSum: true}
	gotChecked := map[string]bool{}
	for _, k := range srv.checked {
		gotChecked[k] = true
	}
	for k := range wantChecked {
		if !gotChecked[k] {
			t.Errorf("object %q was not existence-checked", k)
		}
	}

	// The upload is logged with the manifest name, not just the digest.
	if !strings.Contains(logBuf.String(), "nextcloud33.tar.gz") {
		t.Errorf("log %q does not contain image name", logBuf.String())
	}
}

// writeRawTar writes a tarball with the given entries in the given order, so
// tests can construct malformed tarballs that writeObjects would never emit.
func writeRawTar(t *testing.T, path string, entries [][2]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tar: %v", err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	for _, e := range entries {
		name, body := e[0], e[1]
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Mode:     0644,
			Size:     int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header %q: %v", name, err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("write body %q: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
}

func manifestJSON(t *testing.T, names map[string]string) string {
	t.Helper()
	bs, err := json.Marshal(&objectManifest{Names: names})
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	return string(bs)
}

func TestUploaderUploadManifestNotFirst(t *testing.T) {
	dir := t.TempDir()
	objs := filepath.Join(dir, "objs.tar")
	writeRawTar(t, objs, [][2]string{
		{"fake:aaa", "content-a"},
		{objectManifestName, manifestJSON(t, map[string]string{
			"fake:aaa": "core.tar.gz",
		})},
	})

	srv := newFakeObjServer(nil)
	defer srv.close()

	err := srv.uploader().Upload(objs)
	if err == nil {
		t.Fatalf("Upload got nil error, want manifest-not-first error")
	}
	if !strings.Contains(err.Error(), "first file") {
		t.Errorf("error = %v, want mention of first file", err)
	}
	if len(srv.uploaded) != 0 {
		t.Errorf("uploaded %d objects, want 0", len(srv.uploaded))
	}
}

func TestUploaderUploadMissingName(t *testing.T) {
	dir := t.TempDir()
	objs := filepath.Join(dir, "objs.tar")
	// Manifest is first, but it does not name fake:bbb.
	writeRawTar(t, objs, [][2]string{
		{objectManifestName, manifestJSON(t, map[string]string{
			"fake:aaa": "core.tar.gz",
		})},
		{"fake:bbb", "content-b"},
	})

	srv := newFakeObjServer(nil)
	defer srv.close()

	err := srv.uploader().Upload(objs)
	if err == nil {
		t.Fatalf("Upload got nil error, want missing-name error")
	}
	if !strings.Contains(err.Error(), "no name") {
		t.Errorf("error = %v, want mention of missing name", err)
	}
}

func TestUploaderUploadNoManifest(t *testing.T) {
	dir := t.TempDir()
	objs := filepath.Join(dir, "objs.tar")
	writeRawTar(t, objs, nil) // empty tarball, no manifest

	srv := newFakeObjServer(nil)
	defer srv.close()

	err := srv.uploader().Upload(objs)
	if err == nil {
		t.Fatalf("Upload got nil error, want no-manifest error")
	}
	if !strings.Contains(err.Error(), "no manifest") {
		t.Errorf("error = %v, want mention of missing manifest", err)
	}
}

func TestUploaderUploadPutError(t *testing.T) {
	dir := t.TempDir()
	objs, _ := makeObjsTar(t, dir, map[string]string{"core.tar.gz": "content-a"})

	srv := newFakeObjServer(nil) // nothing exists, so it will try to PUT
	srv.putFail = true
	defer srv.close()

	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	if err := srv.uploader().Upload(objs); err == nil {
		t.Fatalf("Upload got nil error, want upload failure")
	}
}
