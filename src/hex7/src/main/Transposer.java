package main;

import hex7.Hex7Torus;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

/**
 * Reads a {@code MxN} torus representation (hexagon orientation ⬢) from the standard input and
 * prints it out as a {@code N×M} torus (hexagon orientation ⬣).
 * <p>
 * Example of the supported format:
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
public class Transposer {

  public static void main(String[] args) throws IOException {
    BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));
    Hex7Torus torus = null;//Verifier.readTorus(reader);

    System.out.printf("Read %d × %d ⬢ torus successfully.\n", torus.width(), torus.height());

    System.out.printf("\nPrinting as a ⬣ torus:\n");

    for (int line = 0; line < 2 * torus.width(); line++) {
      int start = line % 2 == 0 ? (line / 2 + 1) : (torus.width() + line / 2 + 1);
      start = 2 * torus.width() - start;
      if (line % 2 == 0) {
        System.out.print(" ");
      }
      for (int i = 0; i < torus.height() / 2; i++) {
        int index = start + (i * 2 * torus.width());
        System.out.printf("%s ", torus.get(index).value() == 0x00 ? "0" : "1");
      }
      System.out.println();
    }
  }
}
