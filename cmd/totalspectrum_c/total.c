// comparison C code, as a benchmark
#include <stdio.h>
#include <stdlib.h>

enum { num_channels = 4096 };

int main(int argc, char *argv[]) {
  if (argc != 3) {
    printf("Usage: %s gammafile outfile\n", argv[0]);
    exit(1);
  }
  FILE *gammafp = fopen(argv[1], "r");
  FILE *outfp = fopen(argv[2], "w");
  fprintf(outfp, "total\n");
  char buf[1024];
  fgets(buf, sizeof(buf), gammafp); // skip header line
  int board, channel, energy, energyshort, flags;
  long timetag;
  double total[num_channels] = {0};
  double tmax = 0;
  int count;
  while ((count = fscanf(gammafp, "%d;%d;%ld;%d;%d;%x", &board, &channel,
                         &timetag, &energy, &energyshort, &flags)) != EOF) {
    if (count != 6) {
      printf("Bad count: %d\n", count);
      exit(1);
    }
    total[energy] += 1;
    tmax = timetag / 1.0e12;
  }
  printf("tmax: %f\n", tmax);
  for (int i = 0; i < num_channels; i++) {
    fprintf(outfp, "%f\n", total[i] / tmax);
  }
  fclose(outfp);
  fclose(gammafp);
}
