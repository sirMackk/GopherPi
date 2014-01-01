gopi_media (beta)
=================

A Golang1.1, Raspberry Pi and Firefox-VLC web plugin powered streaming media server. For those of us too convenient to set up a NAS or preposterously use physical devices to access their media across computers. 

Just plug in a hard drive to your RPi, add folder paths and bam, enjoy a web accessible and searchable list of your movies and music. Click on and it'll use the Firefox VLC web plugin to play that media.

### Notes

**Warning** I haven't yet got this working on my RPi due to issues with sqlite on that platform. However, it runs pretty good on debian based distros. 

I had a choice to make: play any media by having the RPi recode it into a web friendly format like mp4 or use a plugin to play any media without the extra work. I chose the Firefox-only VLC web plugin because it plays everything from .avi to .mkv great and it's easily customizable. See *possible improvements*.

### TODO list and possible improvements

Things on my todo list now:

1. Repair the routes, redirects and http errors. Fix template hard coded routes.
2. Buildout a better error handling package.
3. ~~Build a better way to discern between json and non-json requests (or use 3rd party package for that, but then where's the fun?)~~ (12/16/13 - decided against to reduce complexity)
4. ~~Expand the angular SPA to all areas of the site, including user media and admin panel.~~ (12/16/13 - Same as above)
5. Actually add valid and good looking markup instead of the current *hey, lemme see if it works* pile of. Yeah.
6. Add video and audio controls like text/audio subtitle control. (1/1/14 - working on it)
7. Actually get this working on the Raspberry Pi
8. Add config file way of initializing the server
9. Log user interactions ie. logins, adding of files

A little later:

1. Add awesome support for audio files - ~~editable playlists in angular would be awesome.~~ (12/16/13 - Do this in either angular or simple jquery)
2. A way to categorize music albums in some way (tags, folders?)
3. Add fetching of metadata ie. lyrics, imdb ratings/info, duration for music and video.

Possible improvements:

1. Add possibility to transcode video and allow for multiple versions of media to exist - this would allow to view media on any device that supports web video formats (great for watching videos on a tablet at home).

### Security Bonus
This whole application is a huge experiment in learning Golang and building certain core components of a web application by hand. Part of this is the authentication/authorization system. Some parts were knowingly simplified (password storage), whilst some were likely overlooked so bear in mind that **this application is not suitable being for being accessible on the Internet**. Keep it on your LAN. Don't forward ports to it. Play it safe.

#### License

Copyright (C) 2013 sirMackk

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.
