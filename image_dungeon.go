package random

import (
	. "./dungeon"
	"code.google.com/p/draw2d/draw2d"
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

const (
	GRID_NONE   = 0
	GRID_HEX    = 1
	GRID_SQUARE = 2
)

type DrawerConfig struct {
	Fill uint //'000000',
	Grid uint //'CC0000',
	//Open     uint //'FFFFFF',
	OpenGrid uint //'CCCCCC',
	Stairs   uint
	Arch     uint
	Door     uint

	Room      uint
	Corridor  uint
	Perimeter uint

	Labels uint

	CellSize int
	GridType uint
}

type Image struct {
	//cell_size int
	*DrawerConfig
	*image.RGBA
}

func NewDrawer(config *DrawerConfig) (img *Image) {
	return &Image{DrawerConfig: config}
}

func (img *Image) Draw(dungeon *Dungeon) draw.Image {
	if img.CellSize == 0 {
		panic("zero CellSize")
	}
	rect := image.Rect(0, 0,
		(dungeon.Width()+1)*img.CellSize,
		(dungeon.Height()+1)*img.CellSize)
	img.RGBA = image.NewRGBA(rect)

	img.fill_image()
	img.open_cells(dungeon)
	img.image_walls(dungeon, 0x00CC00)
	img.image_doors(dungeon)
	img.image_labels(dungeon)
	img.image_stairs(dungeon)

	return img.RGBA
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
	img.rectangle(0, 0, max.X, max.Y, img.Fill)

	switch img.GridType {
	case GRID_HEX:
		img.grid_hex(img.Grid)
	case GRID_SQUARE:
		img.grid_square(img.Grid)
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
	dim := float64(img.CellSize)
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

	dim := float64(img.CellSize)
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

func (img *Image) open_cells(dungeon *Dungeon) {
	dim := img.CellSize

	//my $base = $image->{'base_layer'};

	for y := 0; y < dungeon.Height(); y++ {
		for x := 0; x < dungeon.Width(); x++ {
			cell := dungeon.At(x, y)
			x1 := x * dim
			y1 := y * dim

			switch {
			case cell&ROOM != 0:
				img.rectangleWH(x1, y1, dim+1, dim+1, img.Room)
			case cell&CORRIDOR != 0:
				img.rectangleWH(x1, y1, dim+1, dim+1, img.Corridor)
			case cell&PERIMETER != 0:
				img.rectangleWH(x1, y1, dim+1, dim+1, img.Perimeter)

			}
		}
	}
}

func (img *Image) image_walls(dungeon *Dungeon, c uint) {
	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(c))
	gc.SetLineWidth(2.0)

	dim := float64(img.CellSize)

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

func (img *Image) image_doors(dungeon *Dungeon) {

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

	dim := float64(img.CellSize)
	a_px := dim / 6.0
	d_tx := dim / 4.0
	t_tx := dim / 3.0
	//pal = $image->{'palette'};

	arch_color := rgba(img.Arch)
	door_color := rgba(img.Arch)

	gc := draw2d.NewGraphicContext(img.RGBA)
	defer gc.Stroke()

	for _, door := range list {
		r := door.Y()
		y1 := float64(r) * dim
		y2 := y1 + dim
		c := door.X()
		x1 := float64(c) * dim
		x2 := x1 + dim

		xc, yc := float64(0), float64(0)
		if dungeon.At(c-1, r)&OPENSPACE != 0 {
			xc = (x1 + x2) / 2.0
		} else {
			yc = (y1 + y2) / 2.0
		}
		attr := door_attr(door)

		gc.SetStrokeColor(arch_color)
		gc.SetLineWidth(3.0)
		if attr.wall {
			if xc != 0 {
				gc.MoveTo(xc, y1)
				gc.LineTo(xc, y2)
			} else {
				gc.MoveTo(x1, yc)
				gc.LineTo(x2, yc)
			}
		}
		gc.Stroke()

		gc.SetStrokeColor(door_color)
		gc.SetLineWidth(1.5)
		if attr.secret {
			if xc != 0 {
				yc := (y1 + y2) / 2
				gc.MoveTo(xc-1, yc-d_tx)
				gc.LineTo(xc+2, yc-d_tx)

				gc.MoveTo(xc-2, yc-d_tx+1)
				gc.LineTo(xc-2, yc-1)

				gc.MoveTo(xc-1, yc)
				gc.LineTo(xc+1, yc)

				gc.MoveTo(xc+2, yc+1)
				gc.LineTo(xc+2, yc+d_tx-1)

				gc.MoveTo(xc-2, yc+d_tx)
				gc.LineTo(xc+1, yc+d_tx)
			} else {
				xc := (x1 + x2) / 2
				gc.MoveTo(xc-d_tx, yc-2)
				gc.LineTo(xc-d_tx, yc+1)

				gc.MoveTo(xc-d_tx+1, yc+2)
				gc.LineTo(xc-1, yc+2)

				gc.MoveTo(xc, yc-1)
				gc.LineTo(xc, yc+1)

				gc.MoveTo(xc+1, yc-2)
				gc.LineTo(xc+d_tx-1, yc-2)

				gc.MoveTo(xc+d_tx, yc-1)
				gc.LineTo(xc+d_tx, yc+2)
			}
		}
		gc.Stroke()

		gc.SetStrokeColor(arch_color)
		gc.SetLineWidth(1.5)
		if attr.arch {
			if xc != 0 {
				draw2d.Rect(gc, xc-1, y1, xc+1, y1+a_px)
				draw2d.Rect(gc, xc-1, y2-a_px, xc+1, y2)
			} else {
				draw2d.Rect(gc, x1, yc-1, x1+a_px, yc+1)
				draw2d.Rect(gc, x2-a_px, yc-1, x2, yc+1)
			}
		}
		gc.Stroke()

		gc.SetStrokeColor(door_color)
		gc.SetLineWidth(1.5)
		if attr.door {
			if xc != 0 {
				draw2d.Rect(gc, xc-d_tx, y1+a_px+1, xc+d_tx, y2-a_px-1)
			} else {
				draw2d.Rect(gc, x1+a_px+1, yc-d_tx, x2-a_px-1, yc+d_tx)
			}
		}

		if attr.lock {
			if xc != 0 {
				gc.MoveTo(xc, y1+a_px+1)
				gc.LineTo(xc, y2-a_px-1)
			} else {
				gc.MoveTo(x1+a_px+1, yc)
				gc.LineTo(x2-a_px-1, yc)
			}
		}

		if attr.trap {
			if xc != 0 {
				yc = (y1 + y2) / 2
				gc.MoveTo(xc, y1+t_tx+1)
				gc.LineTo(xc, y2-t_tx-1)
			} else {
				xc = (x1 + x2) / 2
				gc.MoveTo(x1+t_tx+1, yc)
				gc.LineTo(x2-t_tx-1, yc)
			}
		}
		gc.Stroke()

		gc.SetLineWidth(0.5)
		if attr.portc {
			if xc != 0 {
				start, end := y1+a_px+2, y2-a_px
				gc.MoveTo(xc, start)
				for y := start; y < end; y += 2.0 {
					gc.LineTo(xc, y)
				}
			} else {
				start, end := x1+a_px+2, x2-a_px
				gc.MoveTo(start, yc)
				for x := start; x < end; x += 2.0 {
					gc.LineTo(x, yc)
				}
			}
		}

		gc.Stroke()
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
	switch door.Key() {
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
		attr.arch = true
		attr.portc = true
	}
	return
}

func (img *Image) image_labels(dungeon *Dungeon) {
	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetFillColor(rgba(img.Labels))
	gc.SetFontSize(10.0)
	draw2d.SetFontFolder(".")

	dim := img.CellSize
	for r := 0; r < dungeon.Height(); r++ {
		for c := 0; c < dungeon.Width(); c++ {
			cell := dungeon.At(c, r)
			char := rune((cell >> 24) & 0xFF)
			if cell&OPENSPACE == 0 || char == 0 {
				continue
			}
			s := string([]rune{char})
			left, top, right, bottom := gc.GetStringBounds(s)
			w, h := right-left, bottom-top
			x := float64(c*dim) + w
			y := float64(r*dim) + h
			gc.FillStringAt(s, x, y)
		}
	}
}

func (img *Image) image_stairs(dungeon *Dungeon) {
	list := dungeon.Stairs()
	if len(list) == 0 {
		return
	}

	gc := draw2d.NewGraphicContext(img.RGBA)
	gc.SetStrokeColor(rgba(img.Stairs))
	gc.SetLineWidth(1.0)

	dim := float64(img.CellSize)
	s_px := dim / 2
	t_px := dim/20 + 2

	for _, stair := range list {
		switch {
		case stair.NextY() > stair.Y():
			xc := (float64(stair.X()) + 0.5) * dim
			y1 := float64(stair.Y()) * dim
			y2 := float64(stair.NextY()+1) * dim
			for y := y1; y < y2; y += t_px {
				dx := s_px
				if stair.IsDown() {
					dx = (y - y1) / (y2 - y1) * s_px
				}
				gc.MoveTo(xc-dx, y)
				gc.LineTo(xc+dx, y)
			}
		case stair.NextY() < stair.Y():
			xc := (float64(stair.X()) + 0.5) * dim
			y1 := float64(stair.Y()+1) * dim
			y2 := float64(stair.NextY()) * dim
			for y := y1; y > y2; y -= t_px {
				dx := s_px
				if stair.IsDown() {
					dx = (y - y1) / (y2 - y1) * s_px
				}
				gc.MoveTo(xc-dx, y)
				gc.LineTo(xc+dx, y)
			}
		case stair.NextX() > stair.X():
			x1 := float64(stair.X()) * dim
			x2 := float64(stair.NextX()+1) * dim
			yc := (float64(stair.Y()) + 0.5) * dim
			for x := x1; x < x2; x += t_px {
				dy := s_px
				if stair.IsDown() {
					dy = (x - x1) / (x2 - x1) * s_px
				}
				gc.MoveTo(x, yc-dy)
				gc.LineTo(x, yc+dy)
			}
		case stair.NextX() < stair.X():
			x1 := float64(stair.X()+1) * dim
			x2 := float64(stair.NextX()) * dim
			yc := (float64(stair.Y()) + 0.5) * dim
			for x := x1; x > x2; x -= t_px {
				dy := s_px
				if stair.IsDown() {
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
