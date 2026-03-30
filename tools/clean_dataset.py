import os
import shutil
from PIL import Image

path = './data/processed/mushroom/'

# get ride of species with less than 10 photos
for root, dirs, files in os.walk(path):
    for species_dir in dirs:
        length = len(os.listdir(path + species_dir))
        if length < 10:
            shutil.rmtree(path + species_dir)
        

# get ride of corrupted images
for root, dirs, files in os.walk(path):
    for species_dir in dirs:
        for file in os.listdir(path + species_dir):
            try:
                img = Image.open(path + species_dir + '/' + file)
            except OSError:
                os.remove(path + species_dir + '/' + file)