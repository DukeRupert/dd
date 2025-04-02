export type Albums = Album[]

export interface Album {
  artist_id: string
  collectionId: string
  collectionName: string
  cover_image: string
  created: string
  format: string
  genre_id: string
  id: string
  is_favorite: boolean
  label_id: string
  limited_edition: boolean
  notes: string
  purchase_date: string
  purchase_location: string
  purchase_price: number
  release_year: number
  rpm: string
  section_id: string
  title: string
  updated: string
  expand: Expand
}

export interface Expand {
  artist_id: ArtistId
}

export interface ArtistId {
  collectionId: string
  collectionName: string
  country: string
  created: string
  id: string
  name: string
  notes: string
  updated: string
}