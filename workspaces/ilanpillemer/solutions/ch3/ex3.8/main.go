package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/png"
	"log"
	"math"
	"math/big"
	"math/cmplx"
	"os"
)

var v = flag.Bool("v", false, "print verbose logs for debugging")
var c = flag.String("c", "websafe", "show in colours")
var scale = flag.Float64("s", 1, "zoom")
var left = flag.Float64("l", 0, "left shift")
var down = flag.Float64("d", 0, "down shift")
var fractal = flag.String("f", "mandel", "fractal type")

type Grey = color.Gray

var roots = make(map[complex128]int)

var key = -1
var colours = [4]color.RGBA{
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{255, 255, 0, 255},
}

func main() {
	const (
		//		cntr          = -1.45 - 0i
		//https://commons.wikimedia.org/wiki/File:Zoom_around_principal_Misiurewicz_point_for_periods_from_2_to_1024.gif
		// https://en.wikipedia.org/wiki/Misiurewicz_point
		//cntr = -0.77568377 + 0.13646737i // very pretty - but not precise enough
		//cntr = 0.3663629834227643 + 0.5915337732614452i // precise enough
		//cntr = 0.4245127190500396 + 0.2075302281667453i // precise enough
		//cntr          = -1.54368901269109 + 0i
		//cntr          = -0.260143 + 0.6337i //pretty not precise enough
		//cntr = 0.2759353624416824 + 0.0069166138017372i //very pretty buf goes down a hole
		//cntr          = 0.2787724591293833 + 0.0081245796484104i //very pretty buf goes down a hole
		//cntr          = 0.3115076602815077 + 0.0271737195013418i //very pretty buf goes down a hole
		//	cntr          = -0.1010963638456221 + 0.9562865108091415i // precise enough
		//cntr          = 0.2501502296489224 + 0.0000029308049747i // nice but goes off side
		//cntr          = 0.4379242413594628 + 0.3418920843381161i //precise enough
		cntr = -1i // good and precise enough so lets use this one
		//cntr          = -0.75 + 0i
		//cntr          = -1.401155 + 0i
		//cntr          = -0.75 + 0.1i // sea horse valley
		//cntr          = 0.3 + 0.1i // elephant valley
		ymin, ymax    = -2, +2
		xmin, xmax    = -2, +2
		width, height = 128, 128
	)
	flag.Parse()

	var fn128 func(z complex128) color.Color
	var fn64 func(z complex64) color.Color
	//var fnFloat func(rl float64, mg float64) color.Color
	var fnBigFloat func(rl *big.Float, mg *big.Float) color.Color
	var fnBigRat func(rl *big.Rat, mg *big.Rat) color.Color

	switch *fractal {
	case "newton":
		fn128 = newton
	default:
		fn128 = mandelbrot128
		fn64 = mandelbrot64
		//fnFloat = mandelbrotFloat64
		fnBigFloat = mandelbrotBigFloat
		fnBigRat = mandelbrotBigRat

	}

	img := image.NewRGBA(image.Rect(0, 0, width*2, height*2))
	//img := compImg.SubImage(image.Rect(0, 0, width, height))
	for py := 0; py < height; py++ {
		// project py onto domain using y
		y := (float64(py)/height*(ymax-ymin))/(*scale) + ymin/(*scale)
		for px := 0; px < width; px++ {
			x := float64(px)/width*(xmax-xmin)/(*scale) + xmin/(*scale)
			z := complex(x, y)
			img.Set(px, py, fn128(cntr+z))
			img.Set(px+width, py, fn64(complex64(cntr+z)))
			//img.Set(px, py+height, fnFloat(real(cntr)+x, imag(cntr)+y))
			img.Set(px, py+height, fnBigFloat(big.NewFloat(real(cntr)+x), big.NewFloat(imag(cntr)+y)))
			img.Set(px+width, py+height, fnBigRat(new(big.Rat).SetFloat64(x), new(big.Rat).SetFloat64(y)))
		}
	}

	verbosef("# of roots: %d\n", len(roots))
	if *v {
		for k := range roots {
			verbosef("root is: %v\n", k)
		}
	}
	png.Encode(os.Stdout, img)
}

// next guess is z - ( f(x) / 'f(x) )
func newton(z complex128) color.Color {

	const (
		iter     = 50
		contrast = 20
		dz       = 0.000001
		rnd      = 6
	)

	// next guess is z - ( f(x) / 'f(x) )
	for i := 0; i < iter; i++ {
		nextGuess := (cmplx.Pow(z, 4) - 1) / (4 * cmplx.Pow(z, 3))
		z = z - nextGuess

		if cmplx.Abs(cmplx.Pow(z, 4)-1) < 0+dz {
			if _, ok := roots[round(z, rnd)]; !ok {
				key++
				roots[round(z, rnd)] = key
			}

			switch *c {
			case "grey":
				return Grey{255 - uint8(i)*contrast}
			case "websafe":
				return palette.WebSafe[(255-uint(i)*contrast)%216]
			case "plan9":
				return palette.Plan9[(255-uint(i)*contrast)%255]
			case "primary":
				return shaded(colours[roots[round(z, rnd)]], i, contrast)
			default:
				return palette.WebSafe[(255-uint(i)*contrast)%216]
			}
		}
	}
	return color.Black

}

func shaded(c color.Color, i int, contrast int) color.Color {
	r, g, b, _ := c.RGBA()
	if r != 0 {
		r = uint32(255-(i*contrast)) % 255
	}
	if g != 0 {
		g = uint32(255-(i*contrast)) % 255
	}
	if b != 0 {
		b = uint32(255-(i*contrast)) % 255
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func round(z complex128, r int) complex128 {
	rl := real(z)
	mg := imag(z)

	rl = math.Round(rl * math.Pow10(r))
	mg = math.Round(mg * math.Pow10(r))

	rl = rl / math.Pow10(r)
	mg = mg / math.Pow10(r)

	return complex(rl, mg)
}

func verbosef(format string, args ...interface{}) {
	if *v {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func mandelbrot128(z complex128) color.Color {
	const iterations = 500
	const contrast = 80
	var v complex128
	for n := 0; n < iterations; n++ {
		v = v*v + z
		if c := cmplx.Abs(v); c > 2 {
			return palette.WebSafe[(255-uint(n)*contrast)%216]
		}
	}
	return color.Black
}

func mandelbrot64(z complex64) color.Color {
	const iterations = 500
	const contrast = 80
	var v complex64
	for n := 0; n < iterations; n++ {
		v = v*v + z
		if c := cmplx.Abs(complex128(v)); c > 2 {
			return palette.WebSafe[(255-uint(n)*contrast)%216]

		}
	}
	return color.Black
}

func mandelbrotFloat64(rl float64, mg float64) color.Color {
	const iterations = 500
	const contrast = 80
	var vrl float64
	var vmg float64

	for n := 0; n < iterations; n++ {
		vrl, vmg = (vrl*vrl-vmg*vmg)+rl, (vrl*vmg+vrl*vmg)+mg
		if c := math.Sqrt(vrl*vrl + vmg*vmg); c > 2 {
			return palette.WebSafe[(255-uint(n)*contrast)%216]
		}
	}
	return color.Black
}

func mandelbrotBigFloat(rl *big.Float, mg *big.Float) color.Color {

	const iterations = 500
	const contrast = 80
	var vrl, vmg = big.NewFloat(0), big.NewFloat(0)

	for n := 0; n < iterations; n++ {
		//		vrl, vmg = (vrl*vrl-vmg*vmg)+rl, (vrl*vmg+vrl*vmg)+mg
		a, b, c, d := big.NewFloat(0), big.NewFloat(0), big.NewFloat(0), big.NewFloat(0)
		//log.Println(a, b, c, d)
		a = a.Mul(vrl, vrl)
		b = b.Mul(vmg, vmg)
		c = c.Sub(a, b)
		d = d.Add(c, rl)

		e, f, g := big.NewFloat(0), big.NewFloat(0), big.NewFloat(0)
		e = e.Mul(vrl, vmg)
		f = f.Add(e, e)
		g = g.Add(f, mg)

		vrl = d
		vmg = g

		i, j, k, l := big.NewFloat(0), big.NewFloat(0), big.NewFloat(0), big.NewFloat(4)
		i = i.Mul(vrl, vrl)
		j = j.Mul(vmg, vmg)
		k = k.Add(i, j)
		//log.Println(i, j, k, l, k.Cmp(l))
		if k.Cmp(l) > 0 {
			return palette.WebSafe[(255-uint(n)*contrast)%216]
		}
	}
	return color.Black
}

var cnt = 0

//too slow... dont use!
func mandelbrotBigRat(rl *big.Rat, mg *big.Rat) color.Color {
	const iterations = 500
	const contrast = 80
	var vrl, vmg = big.NewRat(0, 1), big.NewRat(0, 1)

	cnt++

	log.Println("cnt", cnt)
	for n := 0; n < iterations; n++ {
		if cnt == 4028 {
			log.Println("n", n)
		}
		a, b, c, d := big.NewRat(0, 1), big.NewRat(0, 1), big.NewRat(0, 1), big.NewRat(0, 1)
		a = a.Mul(vrl, vrl)
		b = b.Mul(vmg, vmg)
		c = c.Sub(a, b)
		d = d.Add(c, rl)

		e, f, g := big.NewRat(0, 1), big.NewRat(0, 1), big.NewRat(0, 1)
		e = e.Mul(vrl, vmg)
		f = f.Add(e, e)
		g = g.Add(f, mg)

		vrl = d
		vmg = g

		i, j, k, l := big.NewRat(0, 1), big.NewRat(0, 1), big.NewRat(0, 1), big.NewRat(4, 1)
		i = i.Mul(vrl, vrl)
		j = j.Mul(vmg, vmg)
		k = k.Add(i, j)
		if k.Cmp(l) > 0 {
			return palette.WebSafe[(255-uint(n)*contrast)%216]
		}
	}
	return color.Black
}
