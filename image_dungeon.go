package main

import (
	"code.google.com/p/draw2d/draw2d"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
)

func (d *DungeonX) Width() int  { return d.n_cols }
func (d *DungeonX) Height() int { return d.n_rows }
func (d *DungeonX) At(x, y int) uint {
	check := x >= 0 && x < d.Width() && y >= 0 && y < d.Height()
	if !check {
		fmt.Printf("%d/%d %d/%d\n", x, d.Width(), y, d.Height())
		return 0
	}
	return d.cell[y][x]
}
func (d *DungeonX) Doors() []*Door   { return d.door }
func (d *DungeonX) Stairs() []*Stair { return d.stair }

type Dungeon interface {
	Width() int
	Height() int
	At(x, y int) uint
	Doors() []*Door
	Stairs() []*Stair
}

const (
	N = NOTHING
	O = OPENSPACE
	d = DOORSPACE
)

var sd = &DungeonX{
	seed:           1,
	n_cols:         101, // 39, // w
	n_rows:         101, // 39, // h
	dungeon_layout: "None",
	//dungeon_layout:  "Round", // Cross, Box, Round
	room_min:        3,
	room_max:        9,
	room_layout:     "Scattered", // Scattered, Packed
	corridor_layout: CORRIDOR_Bent,
	remove_deadends: 100,
	add_stairs:      20,
	/*
		cell: [][]uint{
			{N, N, N, O, O, O, O, O, O, O},
			{N, O, N, O, O, O, O, O, O, O},
			{O, O, N, N, O, O, O, O, O, O},
			{N, O, O, N, N, N, N, O, O, O},
			{N, O, O, N, N, O, O, O, O, O},
			{N, O, O, N, N, O, O, O, O, O},
			{N, O, d, O, d, O, O, O, O, O},
			{N, O, O, N, N, N, O, O, O, O},
			{N, N, N, N, N, N, O, O, O, O},
			{N, N, N, N, N, N, O, O, O, O},
		},
	*/
}

func init() {
	sd.create_dungeon()
}

func main() {
	//img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img := NewImage(sd.Width()+1, sd.Height()+1, 9)

	img.rectangleWH(0, 0, 20, 50, 0xCC0000)
	img.rectangleWH(img.Rect.Max.X-20, img.Rect.Max.Y-50, 20, 50, 0xCC0000)

	img.fill_image()
	img.open_cells(sd)
	img.image_walls(sd, 0x00CC00)
	img.image_doors(sd)

	// img.image_labels()
	img.image_stairs(sd)

	file, err := os.Create("simple.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}

	exec.Command("feh", "simple.png").Run()
}

const (
	GRID_NONE   = 0
	GRID_HEX    = 1
	GRID_SQUARE = 2
)

type Palete struct {
	Fill     uint //'000000',
	Grid     uint //'CC0000',
	Open     uint //'FFFFFF',
	OpenGrid uint //'CCCCCC',

	GridType uint
}

var StandardPalete = Palete{
	0x000000,
	0xCC0000,
	0xFFFFFF,
	0xCCCCCC,
	GRID_SQUARE,
}

type Image struct {
	cell_size int
	//

	*image.RGBA
}

func NewImage(w, h, cell_size int) (img *Image) {
	rect := image.Rect(0, 0, w*cell_size, h*cell_size)
	img = &Image{
		cell_size: cell_size,
		RGBA:      image.NewRGBA(rect),
	}
	return img
}

func rgba(c uint) color.RGBA {
	return color.RGBA{
		uint8(c >> 16), // r
		uint8(c >> 8),  // g
		uint8(c >> 0),  // b
		0xff,           // a
	}
}

func (img *Image) rectangleWH(x, y, w, h int, c uint) {
	src := &image.Uniform{rgba(c)}
	draw.Draw(img, image.Rect(x, y, x+w, y+h), src, image.ZP, draw.Src)
}
func (img *Image) rectangle(x, y, x2, y2 int, c uint) {
	src := &image.Uniform{rgba(c)}
	draw.Draw(img, image.Rect(x, y, x2, y2), src, image.ZP, draw.Src)
}

func (img *Image) fill_image() {
	//dim := img.cell_size
	//max_x := size.X

	/*
	  my ($dungeon,$image,$ih) = @_;
	  my $max_x = $image->{'max_x'};
	  my $max_y = $image->{'max_y'};
	  my $dim = $image->{'cell_size'};
	  my $pal = $image->{'palette'};
	  my ($color,$tile);

	  if (defined ($tile = $pal->{'fill_pattern'})) {
	    $ih->setTile($tile);
	    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
	  } elsif (defined ($tile = $pal->{'fill_tile'})) {
	    my $r; for ($r = 0; $r <= $dungeon->{'n_rows'}; $r++) {
	      my $c; for ($c = 0; $c <= $dungeon->{'n_cols'}; $c++) {
	        $ih->copy($tile,($c * $dim),($r * $dim),&select_tile($tile,$dim));
	      }
	    }
	  } elsif (defined ($color = $pal->{'fill'})) {
	    $ih->filledRectangle(0,0,$max_x,$max_y,$color);
	  } elsif (defined ($tile = $pal->{'background'})) {
	    $ih->setTile($tile);
	    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
	  } else {
	    $ih->filledRectangle(0,0,$max_x,$max_y,$pal->{'black'});
	    $ih->fill(0,0,$pal->{'black'});
	  }
	*/

	max := img.Bounds().Size()
	img.rectangle(0, 0, max.X, max.Y, StandardPalete.Fill)

	switch StandardPalete.GridType {
	case GRID_HEX:
		img.grid_hex(StandardPalete.Grid)
	case GRID_SQUARE:
		img.grid_square(StandardPalete.Grid)
	}

	/*
	  if ($color = $pal->{'fill_grid'}) {
	    $ih = &image_grid($dungeon,$image,$color,$ih);
	  } elsif ($color = $pal->{'grid'}) {
	    $ih = &image_grid($dungeon,$image,$color,$ih);
	  }
	  return $ih;
	*/
}

/*

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# image dungeon

sub image_dungeon {
  my ($dungeon) = @_;
  my $image = &scale_dungeon($dungeon);

  # - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
  # new image

  my $ih = new GD::Image($image->{'width'},$image->{'height'},1);
  my $pal = &get_palette($image,$ih);
     $image->{'palette'} = $pal;
  my $base = &base_layer($dungeon,$image,$ih);
     $image->{'base_layer'} = $base;

  $ih = &fill_image($dungeon,$image,$ih);
  $ih = &open_cells($dungeon,$image,$ih);
  $ih = &image_walls($dungeon,$image,$ih);
  $ih = &image_doors($dungeon,$image,$ih);
  $ih = &image_labels($dungeon,$image,$ih);

  if ($dungeon->{'stair'}) {
    $ih = &image_stairs($dungeon,$image,$ih);
  }

  # - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
  # write image

  open(OUTPUT,">$dungeon->{'seed'}.gif") and do {
    print OUTPUT $ih->gif();
    close(OUTPUT);
  };
  return "$dungeon->{'seed'}.gif";
}
*/
/*
# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# scale dungeon

sub scale_dungeon {
  my ($dungeon) = @_;

  my $image = {
    'cell_size' => $dungeon->{'cell_size'},
    'map_style' => $dungeon->{'map_style'},
  };
  $image->{'width'}  = (($dungeon->{'n_cols'} + 1)
                     *   $image->{'cell_size'}) + 1;
  $image->{'height'} = (($dungeon->{'n_rows'} + 1)
                     *   $image->{'cell_size'}) + 1;
  $image->{'max_x'}  = $image->{'width'} - 1;
  $image->{'max_y'}  = $image->{'height'} - 1;

  if ($image->{'cell_size'} > 16) {
    $image->{'font'} = gdLargeFont;
  } elsif ($image->{'cell_size'} > 12) {
    $image->{'font'} = gdSmallFont;
  } else {
    $image->{'font'} = gdTinyFont;
  }
  $image->{'char_w'} = $image->{'font'}->width;
  $image->{'char_h'} = $image->{'font'}->height;
  $image->{'char_x'} = int(($image->{'cell_size'}
                     -      $image->{'char_w'}) / 2) + 1;
  $image->{'char_y'} = int(($image->{'cell_size'}
                     -      $image->{'char_h'}) / 2) + 1;

  return $image;
}

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# get palette

sub get_palette {
  my ($image,$ih) = @_;

  my $pal; if ($map_style->{$image->{'map_style'}}) {
    $pal = $map_style->{$image->{'map_style'}};
  } else {
    $pal = $map_style->{'Standard'};
  }
  my $key; foreach $key (keys %{ $pal }) {
    if (ref($pal->{$key}) eq 'ARRAY') {
      $pal->{$key} = $ih->colorAllocate(@{ $pal->{$key} });
    } elsif (-f $pal->{$key}) {
      my $tile; if ($tile = new GD::Image($pal->{$key})) {
        $pal->{$key} = $tile;
      } else {
        delete $pal->{$key};
      }
    } elsif ($pal->{$key} =~ /([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})/i) {
      $pal->{$key} = $ih->colorAllocate(hex($1),hex($2),hex($3));
    }
  }
  unless (defined $pal->{'black'}) {
    $pal->{'black'} = $ih->colorAllocate(0,0,0);
  }
  unless (defined $pal->{'white'}) {
    $pal->{'white'} = $ih->colorAllocate(255,255,255);
  }
  return $pal;
}

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# get color

sub get_color {
  my ($pal,$key) = @_;

  while ($key) {
    return $pal->{$key} if (defined $pal->{$key});
    $key = $color_chain->{$key};
  }
  return undef;
}

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# select tile

sub select_tile {
  my ($tile,$dim) = @_;
  my $src_x = int(rand(int($tile->width / $dim))) * $dim;
  my $src_y = int(rand(int($tile->height / $dim))) * $dim;

  return ($src_x,$src_y,$dim,$dim);
}

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# base layer

sub base_layer {
  my ($dungeon,$image,$ih) = @_;
  my $max_x = $image->{'max_x'};
  my $max_y = $image->{'max_y'};
  my $dim = $image->{'cell_size'};
  my $pal = $image->{'palette'};
  my ($color,$tile);

  if (defined ($tile = $pal->{'open_pattern'})) {
    $ih->setTile($tile);
    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
  } elsif (defined ($tile = $pal->{'open_tile'})) {
    my $r; for ($r = 0; $r <= $dungeon->{'n_rows'}; $r++) {
      my $c; for ($c = 0; $c <= $dungeon->{'n_cols'}; $c++) {
        $ih->copy($tile,($c * $dim),($r * $dim),&select_tile($tile,$dim));
      }
    }
  } elsif (defined ($color = $pal->{'open'})) {
    $ih->filledRectangle(0,0,$max_x,$max_y,$color);
  } elsif (defined ($tile = $pal->{'background'})) {
    $ih->setTile($tile);
    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
  } else {
    $ih->filledRectangle(0,0,$max_x,$max_y,$pal->{'white'});
    $ih->fill(0,0,$pal->{'white'});
  }
  if ($color = $pal->{'open_grid'}) {
    $ih = &image_grid($dungeon,$image,$color,$ih);
  } elsif ($color = $pal->{'grid'}) {
    $ih = &image_grid($dungeon,$image,$color,$ih);
  }
  my $base = $ih->clone();

  if (defined ($tile = $pal->{'background'})) {
    $ih->setTile($tile);
    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
  } else {
    $ih->filledRectangle(0,0,$max_x,$max_y,$pal->{'white'});
  }
  return $base;
}

*/

func (img *Image) grid_square(c uint) {
	dim := float64(img.cell_size)
	off := float64(0.5)
	size := img.Bounds().Size()
	max_x := float64(size.X) + off
	max_y := float64(size.Y) + off

	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(c))

	for x := off; x <= max_x; x += dim {
		gc.MoveTo(x, 0)
		gc.LineTo(x, max_y)
	}

	for y := off; y <= max_y; y += dim {
		gc.MoveTo(0, y)
		gc.LineTo(max_x, y)
	}
	gc.Stroke()
}
func (img *Image) grid_hex(c uint) {
	//off := float64(0.5)
	size := img.Bounds().Size()

	dim := float64(img.cell_size)
	dx := dim / 3.4641016151
	dy := dim / 2.0
	col := int(float64(size.X) / (3.0 * dx))
	row := int(float64(size.Y) / dy)

	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(c))

	for i := 0; i < col; i++ {
		x1 := float64(i) * (3 * dx)
		x2 := x1 + dx
		x3 := x1 + (3 * dx)

		for j := 0; j < row; j++ {
			y1 := float64(j) * dy
			y2 := y1 + dy
			if (i+j)%2 == 0 {
				gc.MoveTo(x1, y1)
				gc.LineTo(x2, y2)
				gc.MoveTo(x2, y2)
				gc.LineTo(x3, y2)
			} else {
				gc.MoveTo(x2, y1)
				gc.LineTo(x1, y2)
			}
		}
	}
	gc.Stroke()
}

/*
# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# fill dungeon image

sub fill_image {
  my ($dungeon,$image,$ih) = @_;
  my $max_x = $image->{'max_x'};
  my $max_y = $image->{'max_y'};
  my $dim = $image->{'cell_size'};
  my $pal = $image->{'palette'};
  my ($color,$tile);

  if (defined ($tile = $pal->{'fill_pattern'})) {
    $ih->setTile($tile);
    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
  } elsif (defined ($tile = $pal->{'fill_tile'})) {
    my $r; for ($r = 0; $r <= $dungeon->{'n_rows'}; $r++) {
      my $c; for ($c = 0; $c <= $dungeon->{'n_cols'}; $c++) {
        $ih->copy($tile,($c * $dim),($r * $dim),&select_tile($tile,$dim));
      }
    }
  } elsif (defined ($color = $pal->{'fill'})) {
    $ih->filledRectangle(0,0,$max_x,$max_y,$color);
  } elsif (defined ($tile = $pal->{'background'})) {
    $ih->setTile($tile);
    $ih->filledRectangle(0,0,$max_x,$max_y,gdTiled);
  } else {
    $ih->filledRectangle(0,0,$max_x,$max_y,$pal->{'black'});
    $ih->fill(0,0,$pal->{'black'});
  }
  if (defined ($color = $pal->{'fill'})) {
    $ih->rectangle(0,0,$max_x,$max_y,$color);
  }
  if ($color = $pal->{'fill_grid'}) {
    $ih = &image_grid($dungeon,$image,$color,$ih);
  } elsif ($color = $pal->{'grid'}) {
    $ih = &image_grid($dungeon,$image,$color,$ih);
  }
  return $ih;
}

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# open cells
*/

func (img *Image) open_cells(dungeon Dungeon) {
	dim := img.cell_size

	//my $base = $image->{'base_layer'};

	for y := 0; y < dungeon.Height(); y++ {
		for x := 0; x < dungeon.Width(); x++ {
			if dungeon.At(x, y)&OPENSPACE == 0 {
				continue
			}
			x1 := x * dim
			y1 := y * dim

			img.rectangleWH(x1, y1, dim, dim, StandardPalete.Open)
			//$ih->copy($base,$x1,$y1,$x1,$y1,($dim+1),($dim+1));
		}
	}
}

func (img *Image) image_walls(dungeon Dungeon, c uint) {
	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(c))
	gc.SetLineWidth(2.0)

	dim := float64(img.cell_size)

	for y := 0; y < dungeon.Height(); y++ {
		for x := 0; x < dungeon.Width(); x++ {
			if dungeon.At(x, y)&OPENSPACE == 0 {
				continue
			}

			y1 := float64(y)*dim + 0.5
			y2 := y1 + dim + 0.5
			x1 := float64(x)*dim + 0.5
			x2 := x1 + dim + 0.5

			if dungeon.At(x, y-1)&OPENSPACE == 0 {
				gc.MoveTo(x1, y1)
				gc.LineTo(x2, y1)
			}
			if dungeon.At(x-1, y)&OPENSPACE == 0 {
				gc.MoveTo(x1, y1)
				gc.LineTo(x1, y2)
			}
			if dungeon.At(x, y+1)&OPENSPACE == 0 {
				gc.MoveTo(x1, y2)
				gc.LineTo(x2, y2)
			}
			if dungeon.At(x+1, y)&OPENSPACE == 0 {
				gc.MoveTo(x2, y1)
				gc.LineTo(x2, y2)
			}
		}
	}

	gc.Stroke()
}

func (img *Image) image_doors(dungeon Dungeon) {

	list := dungeon.Doors()
	if len(list) == 0 {
		return
	}

	/*
	  my ($dungeon,$image,$ih) = @_;
	  my $list = $dungeon->{'door'};
	     return $ih unless ($list);

	  my $cell = $dungeon->{'cell'};
	*/

	dim := img.cell_size
	a_px := dim / 6
	d_tx := dim / 4
	//t_tx := dim / 3
	//pal = $image->{'palette'};

	arch_color := uint(0xFF00FF)
	door_color := uint(0x00FFFF)

	gc := draw2d.NewGraphicContext(img.RGBA)

	for _, door := range list {
		fmt.Println("door", door)

		r := door.row
		y1 := r * dim
		y2 := y1 + dim
		c := door.col
		x1 := c * dim
		x2 := x1 + dim

		xc, yc := 0, 0
		if dungeon.At(c-1, r)&OPENSPACE != 0 {
			xc = (x1 + x2) / 2
		} else {
			yc = (y1 + y2) / 2
		}
		attr := door_attr(door)

		gc.SetStrokeColor(rgba(arch_color))
		gc.SetLineWidth(3.0)
		if attr.wall {
			if xc != 0 {
				gc.MoveTo(float64(xc), float64(y1))
				gc.LineTo(float64(xc), float64(y2))
			} else {
				gc.MoveTo(float64(x1), float64(yc))
				gc.LineTo(float64(x2), float64(yc))
			}
		}

		gc.SetStrokeColor(rgba(door_color))
		gc.SetLineWidth(3.0)
		if attr.secret {
			if xc != 0 {
				yc := (y1 + y2) / 2
				gc.MoveTo(float64(xc-1), float64(yc-d_tx))
				gc.LineTo(float64(xc+2), float64(yc-d_tx))

				gc.MoveTo(float64(xc-2), float64(yc-d_tx+1))
				gc.LineTo(float64(xc-2), float64(yc-1))

				gc.MoveTo(float64(xc-1), float64(yc))
				gc.LineTo(float64(xc+1), float64(yc))

				gc.MoveTo(float64(xc+2), float64(yc+1))
				gc.LineTo(float64(xc+2), float64(yc+d_tx-1))

				gc.MoveTo(float64(xc-2), float64(yc+d_tx))
				gc.LineTo(float64(xc+1), float64(yc+d_tx))
			} else {
				xc := (x1 + x2) / 2
				gc.MoveTo(float64(xc-d_tx), float64(yc-2))
				gc.LineTo(float64(xc-d_tx), float64(yc+1))

				gc.MoveTo(float64(xc-d_tx+1), float64(yc+2))
				gc.LineTo(float64(xc-1), float64(yc+2))

				gc.MoveTo(float64(xc), float64(yc-1))
				gc.LineTo(float64(xc), float64(yc+1))

				gc.MoveTo(float64(xc+1), float64(yc-2))
				gc.LineTo(float64(xc+d_tx-1), float64(yc-2))

				gc.MoveTo(float64(xc+d_tx), float64(yc-1))
				gc.LineTo(float64(xc+d_tx), float64(yc+2))
			}
		}

		if attr.arch {
			if xc != 0 {
				img.rectangle(xc-1, y1, xc+1, y1+a_px, arch_color)
				img.rectangle(xc-1, y2-a_px, xc+1, y2, arch_color)
			} else {
				img.rectangle(x1, yc-1, x1+a_px, yc+1, arch_color)
				img.rectangle(x2-a_px, yc-1, x2, yc+1, arch_color)
			}
		}

		if attr.door {
			if xc != 0 {
				img.rectangle(xc-d_tx, y1+a_px+1, xc+d_tx, y2-a_px-1, door_color)
			} else {
				img.rectangle(x1+a_px+1, yc-d_tx, x2-a_px-1, yc+d_tx, door_color)
			}
		}

		/*
		   if ($attr->{'lock'}) {
		     if ($xc) {
		       $ih->line($xc,$y1+$a_px+1,$xc,$y2-$a_px-1,$door_color);
		     } else {
		       $ih->line($x1+$a_px+1,$yc,$x2-$a_px-1,$yc,$door_color);
		     }
		   }
		   if ($attr->{'trap'}) {
		     if ($xc) {
		       my $yc = int(($y1 + $y2) / 2);
		       $ih->line($xc-$t_tx,$yc,$xc+$t_tx,$yc,$door_color);
		     } else {
		       my $xc = int(($x1 + $x2) / 2);
		       $ih->line($xc,$yc-$t_tx,$xc,$yc+$t_tx,$door_color);
		     }
		   }
		   if ($attr->{'portc'}) {
		     if ($xc) {
		       my $y; for ($y = $y1+$a_px+2; $y < $y2-$a_px; $y += 2) {
		         $ih->setPixel($xc,$y,$door_color);
		       }
		     } else {
		       my $x; for ($x = $x1+$a_px+2; $x < $x2-$a_px; $x += 2) {
		         $ih->setPixel($x,$yc,$door_color);
		       }
		     }
		   }
		*/
	}
}

// door attributes

type DoorAttr struct {
	arch   bool
	door   bool
	lock   bool
	trap   bool
	secret bool
	portc  bool
	wall   bool
}

func door_attr(door *Door) (attr DoorAttr) {
	switch door.key {
	case "arch":
		attr.arch = true
	case "open":
		attr.arch = true
		attr.door = true
	case "lock":
		attr.arch = true
		attr.door = true
		attr.lock = true
	case "trap":
		attr.arch = true
		attr.door = true
		// TODO attr.lock = true
		attr.trap = true
	case "secret":
		attr.arch = true
		attr.wall = true
		attr.secret = true
	case "portc":
		fmt.Println("PORTC !!!!!!!!!!!!!!!")
		attr.arch = true
		attr.portc = true
	}
	return
}

/*

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# image labels

sub image_labels {
  my ($dungeon,$image,$ih) = @_;
  my $cell = $dungeon->{'cell'};
  my $dim = $image->{'cell_size'};
  my $pal = $image->{'palette'};
  my $color = &get_color($pal,'label');

  my $r; for ($r = 0; $r <= $dungeon->{'n_rows'}; $r++) {
    my $c; for ($c = 0; $c <= $dungeon->{'n_cols'}; $c++) {
      next unless ($cell->[$r][$c] & $OPENSPACE);

      my $char = &cell_label($cell->[$r][$c]);
         next unless (defined $char);
      my $x = ($c * $dim) + $image->{'char_x'};
      my $y = ($r * $dim) + $image->{'char_y'};

      $ih->string($image->{'font'},$x,$y,$char,$color);
    }
  }
  return $ih;
}

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# cell label

sub cell_label {
  my ($cell) = @_;
  my $i = ($cell >> 24) & 0xFF;
     return unless ($i);
  my $char = chr($i);
     return unless ($char =~ /^\d/);
  return $char;
}
*/

func (img *Image) image_stairs(dungeon Dungeon) {
	list := dungeon.Stairs()
	if len(list) == 0 {
		return
	}

	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(0xff00ff))
	gc.SetLineWidth(1.0)

	dim := float64(img.cell_size)
	s_px := dim / 2
	t_px := dim/20 + 2

	for _, stair := range list {
		fmt.Println("stair", stair)
		switch {
		case stair.next_row > stair.row:
			xc := (float64(stair.col) + 0.5) * dim
			y1 := float64(stair.row) * dim
			y2 := float64(stair.next_row+1) * dim
			for y := y1; y < y2; y += t_px {
				dx := s_px
				if stair.key == "down" {
					dx = (y - y1) / (y2 - y1) * s_px
				}
				gc.MoveTo(xc-dx, y)
				gc.LineTo(xc+dx, y)
			}
		case stair.next_row < stair.row:
			xc := (float64(stair.col) + 0.5) * dim
			y1 := float64(stair.row+1) * dim
			y2 := float64(stair.next_row) * dim
			for y := y1; y > y2; y -= t_px {
				dx := s_px
				if stair.key == "down" {
					dx = (y - y1) / (y2 - y1) * s_px
				}
				gc.MoveTo(xc-dx, y)
				gc.LineTo(xc+dx, y)
			}
		case stair.next_col > stair.col:
			x1 := float64(stair.col) * dim
			x2 := float64(stair.next_col+1) * dim
			yc := (float64(stair.row) + 0.5) * dim
			for x := x1; x < x2; x += t_px {
				dy := s_px
				if stair.key == "down" {
					dy = (x - x1) / (x2 - x1) * s_px
				}
				gc.MoveTo(x, yc-dy)
				gc.LineTo(x, yc+dy)
			}
		case stair.next_col < stair.col:
			x1 := float64(stair.col+1) * dim
			x2 := float64(stair.next_col) * dim
			yc := (float64(stair.row) + 0.5) * dim
			for x := x1; x > x2; x -= t_px {
				dy := s_px
				if stair.key == "down" {
					dy = (x - x1) / (x2 - x1) * s_px
				}
				gc.MoveTo(x, yc-dy)
				gc.LineTo(x, yc+dy)
			}
		default:
			fmt.Println("WUT stair?")
		}
	}
	gc.Stroke()
}
