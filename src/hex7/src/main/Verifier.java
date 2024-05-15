package main;

import hex7.Hex7;
import hex7.Hex7Torus;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.function.Function;

/**
 * Reads a torus representation from the standard input and reports how many unique 7-hexagons are
 * there.
 * <p>
 * Example of the supported format (hexagons are like ⬢):
 * <pre>
 * 0 0 1 1 1 0 1 0 0 1 1 0 1 0 . .
 *  0 0 1 1 1 1 0 0 0 0 0 1 0 1 . .
 * 0 0 0 1 1 0 1 1 1 0 1 0 1 . . .
 *  ( and so on )
 * </pre>
 * Whitespace is ignored, {@code '.'} means the hexagon is unoccupied. The area of the torus is
 * assumed to be 128 hexagons, and its dimensions are determined by the number of elements in the
 * first line.
 */
public class Verifier {

  private static final Function<String, Byte> converter = e ->
      switch (e) {
        case "0", "O" -> (byte) 0;
        case "1", "I" -> (byte) 1;
        default -> (byte) 0xff;
      };

  static Hex7Torus readTorus(BufferedReader reader) throws IOException {
    String line;
    do {
      line = reader.readLine();
    } while (line.strip().split("\\s+").length == 0);

    String[] elements = line.strip().split("\\s+");
    int width = elements.length;
    int height = 22 * 14 / width;

    List<Byte> hexagons = new ArrayList<>();
    Arrays.stream(elements)
        .map(converter)
        .forEach(hexagons::add);

    int read = 1;
    while ((line = reader.readLine()) != null && read < height) {
      elements = line.strip().split("\\s+");
      if (elements.length == 0) {
        continue;
      }
      read++;

      if (elements.length != width) {
        System.out.println(Arrays.toString(elements));
        throw new RuntimeException(
            "Inconsistent line length " + elements.length + " in this line:\n'"
                + line + "'");
      }

      Arrays.stream(elements)
          .map(converter)
          .forEach(hexagons::add);
    }

    Hex7Torus torus = new Hex7Torus(width, height);
    for (int i = 0; i < hexagons.size(); i++) {
      torus.set(i, hexagons.get(i));
    }
    return torus;
  }

  public static void main(String[] args) throws IOException {
    BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
    Hex7Torus torus = readTorus(reader);

    System.out.printf("Read %d × %d ⬢ torus successfully.\n", torus.width(), torus.height());

    Map<Hex7, Integer> counts = torus.verify();
    if (counts.size() == 128) {
      System.out.println("The torus is fully populated.");
    }

    long invalid = counts.entrySet().stream()
        .filter(e -> e.getValue() != 1)
        .count();
    if (invalid != 0) {
      System.out.println("The torus is invalid, these Hex7s appear more than once:");
      counts.entrySet().stream()
          .filter(e -> e.getValue() != 1)
          .forEach(e -> System.out.printf("%d times:%s\n", e.getValue(), e.getKey()));
    } else {
      System.out.println("Each Hex7 appears exactly once.");
    }
  }
}
