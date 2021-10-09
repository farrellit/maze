Maze Generation
=====================

## Running the WebApp in development mode

Run this in the repository root:
```
DEV=true go run .
```

`http://localhost:1801/devui/` will dynamically serve the contents of `webui/` 

Static site will still be available at `webui/` (see below) but these assets will be 
static as of the time Go was built.  

Please note that regardless of development mode, `/` root url redirects to `webui/`.  

## Running in production mode

without `DEV` environment variable set to true, `devui/` will not route.  Instead, `/webui/` 
serves static assets which were embeddedd into the Go executable at build time. 

# How the program works

## Generating Mazes

There is a `MazeCreator` interface that defines a `Fill` function that takes a grid, start, and finish coordinates.  

Each implementation of `MazeCreator` defines an algorithm to fill a maze.  
Right now the only implementation is `WalkingCreator` which implements a simple random generation algorithm:
1. Start at the "Start" position
2. Randomly choose a sqare orthogonally adjacent to the current position that does not itself adjoin a maze passageway.  If no such location is found, choose a random passageway position to use instead of current position (thus creating another "branch" of the maze at that position)
3. If the new position is adjacent to the finish, we're done.

If after some large number of iterations, in the magnitude of 10^3 to 10^4 or so, the finish has not been reached, then  the algorithm "reverse completes" the maze from finish to any passageway *not* part of the reverse completion.  This essentially works backwards from the Finish to a passageway reached randomly from the start, thus "solving" the maze.

This simple algorithm often generates very dense mazes with many dead ends, but it does not guarantee any particular density of maze - it is entirely possible the generator might generate a very simple maze, even an unbifurcated path from start directly to finish.  

## Drawing Mazes

The `Renderer` interface defines a `Draw` function that takes a `*Maze` and draws it out.

There are 2 renderers defined, a `ConsoleRenderer` that writes to a text terminal (tpyically used for debugging), and a `SVGRenderer` that renders an SVG.

## The website

A small web interface handles collecting X and Y dimensions of the maze, a scale (which is more or less irrelevant since the picture is rendered in SVG anyway), and a seed for the API's random number generator, which is randomly set in the javascript side.

When these numbers are changed, a new maze is generated and shown as an image below.   The maze's URL is meant to be persistent - further requests should always generate the same maze since they start with same dimensions and random seed.  Thus the `/api/maze/...` URLs can be cached, and the URLs produce consistent results on subsequent requests.

If a maze is requested without a random seed, the API redirects to a URL with a random seed suppled on the api side.  It also sets defaults for X and Y if none are set, and redirects to a new URL with all that stuff supplied.  Scale also has a default.  See `main.go` for these details.

On print media, the maze controls are styled such that they should not be printed.
