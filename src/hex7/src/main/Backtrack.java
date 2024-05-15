package main;

import hex7.Hex7;
import hex7.Hex7Torus;
import hex7.HexCoord;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.Collections;
import java.util.List;
import java.util.Random;

/**
 * Goes through all possible {@link Hex7Torus} of a certain size using backtracking, and terminates
 * immediately if a fully-populated torus is found.
 */
public class Backtrack {

  private static final Random random = new Random();

  private static final int W = 32;
  private static final int H = 4;

  private static final BitSet visited = new BitSet();
  private static final Hex7Torus torus = new Hex7Torus(W, H);
  private static final List<HexCoord> sequence = new ArrayList<>();
  private static long step = 0;
  private static int max = 50;

  private static void dfs(int depth) {
    HexCoord curr = sequence.get(depth);

    if (++step % 10_000_000 == 0) {
      System.out.printf("\nStep: %,d, max: %d\n%s\n", step, max, torus);
    }

    BitSet available = torus.available(curr);
    available.andNot(visited);
    if (available.isEmpty()) {
      return;
    }

    List<Integer> list = new ArrayList<>();
    available.stream().forEach(list::add);
    Collections.shuffle(list, random);

    list.stream()
        .map(Hex7::of)
        .forEach(hex -> {
          // Set hex7 at 'curr' to one of the options.
          torus.set(curr, hex);
          visited.set(hex.value());

          if (visited.cardinality() > max) {
            max = visited.cardinality();
            System.out.println(
                "\n"
                    + "============================="
                    + "\nBest cardinality: " + max + "\n" + torus
                    + "\n=============================");
          }

          if (visited.cardinality() == 128) {
            System.out.printf("Found the torus (%,d steps):\n%s\n", step, torus);
            System.exit(0);
          }

          dfs(depth + 1);

          // Restore the torus & visited, so the caller can try another option.
          torus.unset(curr);
          visited.clear(hex.value());
        });
  }

  public static void main(String[] args) {
    long seed = random.nextLong();
    System.out.println("Random seed: " + seed);
    random.setSeed(seed);

    for (int x = 0; x < W; x++) {
      for (int y = 0; y < H / 2; y++) {
        sequence.add(HexCoord.of(x, 2 * y));
      }
      for (int y = 0; y < H / 2; y++) {
        sequence.add(HexCoord.of(x, 2 * y + 1));
      }
    }

    Hex7 hex = Hex7.of(random.nextInt(128));
    torus.set(sequence.getFirst(), hex);
    visited.set(hex.value());

    dfs(1);
  }
}
