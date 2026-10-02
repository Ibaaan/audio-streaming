// // Command stripflac removes metadata that is not needed for playback
// // from a FLAC file.
// //
// // Usage: stripflac <in.flac> <out.flac>
package main

// import (
// 	"fmt"
// 	"os"

// 	// "audio-streaming/flac"
// )

// func main() {
// 	if len(os.Args) != 3 {
// 		fmt.Fprintln(os.Stderr, "usage: stripflac <in.flac> <out.flac>")
// 		os.Exit(2)
// 	}

// 	if err := run(os.Args[1], os.Args[2]); err != nil {
// 		fmt.Fprintln(os.Stderr, "stripflac:", err)
// 		os.Exit(1)
// 	}
// }

// func run(inPath, outPath string) (err error) {
// 	if inPath == outPath {
// 		return fmt.Errorf("input and output must differ")
// 	}

// 	in, err := os.Open(inPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() { _ = in.Close() }()

// 	out, err := os.Create(outPath)
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if cerr := out.Close(); err == nil {
// 			err = cerr
// 		}

// 		if err != nil {
// 			_ = os.Remove(outPath)
// 		}
// 	}()

// 	return flac.Strip(out, in)
// }
