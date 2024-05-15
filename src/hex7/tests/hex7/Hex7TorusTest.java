package hex7;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.BitSet;
import java.util.Collections;
import java.util.List;
import java.util.Random;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.ArgumentsProvider;
import org.junit.jupiter.params.provider.ArgumentsSource;

class Hex7TorusTest {

  private static final Random rand = new Random();

  // Randomize the torus - it's not valid, but good enough for test assertions.
  private static Hex7Torus randomizeTorus(Hex7Torus torus) {
    for (int j = 0; j < 200; j++) {
      HexCoord coord = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      if (rand.nextInt(100) < 25) {
        torus.unset(coord);
      } else {
        setRandomHex7(torus, coord);
      }
    }
    return torus;
  }

  private static BitSet getVisited(Hex7Torus torus) {
    BitSet visited = new BitSet();
    for (int i = 0; i < torus.array.length; i++) {
      if (torus.array[i] != (byte) 0xff) {
        visited.set(torus.array[i]);
      }
    }
    return visited;
  }

  private static int calculateUnsetSize(Hex7Torus torus) {
    int result = 0;
    for (int i = 0; i < torus.array.length; i++) {
      if (torus.array[i] == (byte) 0xff) {
        result++;
      }
    }
    return result;
  }

  // Set a random Hex7 at 'coord' such as it doesn't conflict with any other Hex7s around it.
  private static void setRandomHex7(Hex7Torus torus, HexCoord coord) {
    coord = torus.normalize(coord);

    List<Integer> options = new ArrayList<>();
    torus.available(coord)
        .stream()
        .forEach(options::add);
    Collections.shuffle(options);

    if (options.isEmpty()) {
      // There's always an option available, because in the test we allow duplicate Hex7s on a torus
      throw new RuntimeException(String.format("No available options at %s for\n%s", coord, torus));
    }

    torus.set(coord, Hex7.of(options.get(0)));
  }

  @ParameterizedTest
  @ArgumentsSource(TorusProvider.class)
  public void surround_full(int w, int h) {
    // If Hex7s 'a', 'b' and 'c' are set, then Hex7 centered at 'x' has only 2 options.
    //  a a   b b
    // a a a b b b
    //  a a x b b
    //     c c
    //    c c c
    //     c c
    // The centers of 'a', 'b' and 'c' are at UL+L, UR+R and DR+DL.
    for (int i = 0; i < 1; i++) {
      Hex7Torus torus = new Hex7Torus(w, h);

      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      setRandomHex7(torus, x.move(HexDir.UL).move(HexDir.ZL));
      setRandomHex7(torus, x.move(HexDir.UR).move(HexDir.R));
      setRandomHex7(torus, x.move(HexDir.DR).move(HexDir.DL));

      BitSet available = torus.available(x);
      assertEquals(2, available.cardinality(),
          String.format("Fully surrounded Hex7 %s has only 2 options:\n%s", x, torus));
    }
  }

  @Test
  public void surround_knightsMove() {
    for (int i = 0; i < 10_000; i++) {
      Hex7Torus torus = new Hex7Torus(16, 8);
      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      int sides = 0;
      if (rand.nextBoolean()) {
        sides++;
        setRandomHex7(torus, x.move(HexDir.UL).move(HexDir.ZL));
      }
      if (rand.nextBoolean()) {
        sides++;
        setRandomHex7(torus, x.move(HexDir.UR).move(HexDir.R));
      }
      if (rand.nextBoolean()) {
        sides++;
        setRandomHex7(torus, x.move(HexDir.DR).move(HexDir.DL));
      }

      BitSet available = torus.available(x);
      assertEquals(1 << (7 - 2 * sides), available.cardinality(),
          String.format("Hex7 surrounded on %d sizes has only %d options:\n%s", sides,
              1 << (7 - 2 * sides), torus));
    }
  }

  @Test
  public void surround_distance_5() {
    // In this configuration, there's 2^5 possibilities:
    //   a b x x h i
    //  c d E x J k l
    //   f g x x m n
    for (int i = 0; i < 10_000; i++) {
      Hex7Torus torus = new Hex7Torus(16, 8);

      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      HexDir dir = HexDir.values()[rand.nextInt(6)];

      setRandomHex7(torus, x.move(dir).move(dir));
      setRandomHex7(torus, x.move(dir.invert()).move(dir.invert()));

      BitSet available = torus.available(x);
      assertEquals(32, available.cardinality(),
          "Hex7 surrounded by 2 hex7s 5 hexagons apart has 32 options");
    }
  }

  @Test
  public void surround_knights_move_5() {
    for (int i = 0; i < 10_000; i++) {
      Hex7Torus torus = new Hex7Torus(16, 8);

      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      HexDir dir = HexDir.values()[rand.nextInt(6)];
      HexDir dir2 = dir.next().next();

      setRandomHex7(torus, x.move(dir).move(dir.next()));
      setRandomHex7(torus, x.move(dir2).move(dir2));

      BitSet available = torus.available(x);
      assertEquals(16, available.cardinality(),
          "Hex7 surrounded by a knight's move a distance 2 hex7 has 16 options");
    }
  }

  @Test
  public void surround_knights_move_5_shifted() {
    for (int i = 0; i < 10_000; i++) {
      Hex7Torus torus = new Hex7Torus(16, 8);

      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      HexDir dir = HexDir.values()[rand.nextInt(6)];
      HexDir dir2 = dir.invert();

      setRandomHex7(torus, x.move(dir).move(dir.next()));
      setRandomHex7(torus, x.move(dir2).move(dir2));

      assertEquals(16, torus.available(x).cardinality(),
          "Hex7 surrounded by a knight's move a distance 2 hex7 has 16 options.");
    }
  }

  @ParameterizedTest
  @ArgumentsSource(TorusProvider.class)
  public void surround_distance_1(int w, int h) {
    for (int i = 0; i < 10_000; i++) {
      Hex7Torus torus = new Hex7Torus(w, h);

      HexCoord x = HexCoord.of(rand.nextInt(torus.width()), rand.nextInt(torus.height()));
      HexDir dir = HexDir.values()[rand.nextInt(6)];

      setRandomHex7(torus, x.move(dir));

      BitSet available = torus.available(torus.normalize(x.move(dir.invert())));
      if (torus.width() == 16 && torus.height() == 8) {
        // Doesn't hold on 32x4 because it wraps around too tightly.
        assertEquals(64, available.cardinality(),
            "Hex7 that is 3 hexagons from another hex7 has 64 options.");
      }
      setRandomHex7(torus, x.move(dir.invert()));

      available = torus.available(x);
      assertEquals(1, available.cardinality(),
          "Hex7 surrounded by two hex7s at distance 1 has a single option.");
    }
  }

  @ParameterizedTest
  @ArgumentsSource(TorusProvider.class)
  public void populated(int w, int h) {
    Hex7Torus torus = new Hex7Torus(w, h);

    assertEquals(128, torus.calculateUnsetSize(), "Initial torus has 0 hexes populated.");

    HexCoord coord = HexCoord.of(0, 0);
    for (int i = 0; i < 10; i++, coord = coord.move(HexDir.DR)) {
      setRandomHex7(torus, coord);
    }
    assertEquals(118, torus.calculateUnsetSize(), "10 hexes were populated.");

    coord = coord.move(HexDir.UL);
    for (int i = 0; i < 10; i++, coord = coord.move(HexDir.UL)) {
      torus.unset(torus.normalize(coord));
    }
    assertEquals(128, torus.calculateUnsetSize(), "Return to full unpopulated torus.");
  }

  @ParameterizedTest
  @ArgumentsSource(TorusProvider.class)
  public void visited(int w, int h) {
    for (int i = 0; i < 5_000; i++) {
      Hex7Torus torus = randomizeTorus(new Hex7Torus(w, h));

      assertEquals(getVisited(torus), torus.calculateVisited(),
          "Visited sets should be equal: " + torus);
      assertEquals(calculateUnsetSize(torus), torus.calculateUnsetSize(),
          "Unset sizes should be equal: " + torus);
    }
  }

  static class TorusProvider implements ArgumentsProvider {

    @Override
    public Stream<Arguments> provideArguments(ExtensionContext extensionContext) {
      return Stream.of(
          Arguments.arguments(16, 8),
          Arguments.arguments(32, 4)
      );
    }
  }
}

