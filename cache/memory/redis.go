// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package memory


// var preffix = "_PAGE_CACHE_"

// //Storage mecanism for caching strings in memory
// type Storage struct {
// 	client *r.Client
// }

// //NewStorage creates a new redis storage
// func NewStorage(url string) (*Storage, error) {
// 	var (
// 		opts *r.Options
// 		err  error
// 	)

// 	if opts, err = r.ParseURL(url); err != nil {
// 		return nil, err
// 	}

// 	return &Storage{
// 		client: r.NewClient(opts),
// 	}, nil
// }

// //Get a cached content by key
// func (s Storage) Get(key string) []byte {
// 	val, _ := s.client.Get(preffix + key).Bytes()
// 	return val
// }

// //Set a cached content by key
// func (s Storage) Set(key string, content []byte, duration time.Duration) {
// 	s.client.Set(preffix+key, content, duration)
// }
