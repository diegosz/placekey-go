# DEV

## Max distance 2 points in adjacent hexagons

| length (side) | L                 | 65.907807          | 65.907807   | meters
| apothem       | AP=L*SQRT(3)/2    | =D1*SQRT(3)/2      | 57.07783517 |
| shortdiam     | SD=AP*2           | =D2*2              | 114.1556703 |
| perimeter     | P=L*6             | =6*D1              | 395.446842  |
| area          | A=SD*SD*SQRT(3)/2 | =D3*D3*SQRT(3)/2   | 11285.62483 |
| area          | A=P*AP/2          | =D4*D2/2           | 11285.62483 |
| length/2      | B=L/2             | =D1/2              | 32.9539035  |
| h             | H=SQRT(SD*SD+B*B) | =SQRT(D3*D3+D7*D7) | 118.8169888 |
| maxdist       | MD=H*2            | =D8*2              | 237.6339776 | meters

### GPS  accuracy

decimal
places  degrees      N/S or E/W     E/W at         E/W at       E/W at
                     at equator     lat=23N/S      lat=45N/S    lat=67N/S
------- -------      ----------     ----------     ---------    ---------
0       1            111.32 km      102.47 km      78.71 km     43.496 km
1       0.1          11.132 km      10.247 km      7.871 km     4.3496 km
2       0.01         1.1132 km      1.0247 km      787.1 m      434.96 m
3       0.001        111.32 m       102.47 m       78.71 m      43.496 m
4       0.0001       11.132 m       10.247 m       7.871 m      4.3496 m
5       0.00001      1.1132 m       1.0247 m       787.1 mm     434.96 mm
6       0.000001     11.132 cm      102.47 mm      78.71 mm     43.496 mm
7       0.0000001    1.1132 cm      10.247 mm      7.871 mm     4.3496 mm
8       0.00000001   1.1132 mm      1.0247 mm      0.7871mm     0.43496mm
