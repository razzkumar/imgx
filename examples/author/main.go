package main

import (
	"fmt"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	fmt.Println("=== Author Override Examples ===\n")

	// Example 1: Default author (razzkumar)
	fmt.Println("1. Default author:")
	img1, err := imgx.Load("testdata/branch_flip_horizontal.jpg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Author: %s\n", img1.GetMetadata().Author)
	fmt.Printf("   Software: %s v%s\n\n", img1.GetMetadata().Software, img1.GetMetadata().Version)

	// Example 2: Override author at load time
	fmt.Println("2. Override at load time:")
	img2, err := imgx.Load("testdata/branch_flip_horizontal.jpg", imgx.WithAuthor("John Doe"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Author: %s\n", img2.GetMetadata().Author)
	fmt.Printf("   Software: %s v%s (unchanged)\n\n", img2.GetMetadata().Software, img2.GetMetadata().Version)

	// Example 3: Override author after loading
	fmt.Println("3. Override after loading:")
	img3, err := imgx.Load("testdata/branch_flip_horizontal.jpg")
	if err != nil {
		log.Fatal(err)
	}
	img3.SetAuthor("Jane Smith")
	fmt.Printf("   Author: %s\n", img3.GetMetadata().Author)
	fmt.Printf("   Software: %s v%s (unchanged)\n\n", img3.GetMetadata().Software, img3.GetMetadata().Version)

	// Example 4: Set global default author
	fmt.Println("4. Set global default author:")
	imgx.SetDefaultAuthor("Global Artist")
	img4, err := imgx.Load("testdata/branch_flip_horizontal.jpg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Author: %s\n", img4.GetMetadata().Author)
	fmt.Printf("   Software: %s v%s (unchanged)\n\n", img4.GetMetadata().Software, img4.GetMetadata().Version)

	// Example 5: Per-image override takes precedence over global
	fmt.Println("5. Per-image override > global:")
	img5, err := imgx.Load("testdata/branch_flip_horizontal.jpg", imgx.WithAuthor("Specific Artist"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Author: %s\n", img5.GetMetadata().Author)
	fmt.Printf("   Software: %s v%s (unchanged)\n\n", img5.GetMetadata().Software, img5.GetMetadata().Version)

	// Example 6: Using environment variable
	fmt.Println("6. Environment variable (set IMGX_DEFAULT_AUTHOR=EnvArtist):")
	fmt.Println("   Run: IMGX_DEFAULT_AUTHOR=\"EnvArtist\" go run example_author.go")

	fmt.Println("\n=== Summary ===")
	fmt.Println("Priority order for author:")
	fmt.Println("  1. WithAuthor() load option (per-image)")
	fmt.Println("  2. SetAuthor() method (after load)")
	fmt.Println("  3. SetDefaultAuthor() (global)")
	fmt.Println("  4. IMGX_DEFAULT_AUTHOR env var")
	fmt.Println("  5. Default 'razzkumar'")
	fmt.Println("\nNote: creator_tool (Software + Version) is ALWAYS 'imgx v1.0.0'")
}
