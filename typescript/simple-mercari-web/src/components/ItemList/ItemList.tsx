import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image: string;
};


const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';
const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([])
  const fetchItems = () => {
    fetch(server.concat('/items'),
      {
        method: 'GET',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
      })
      .then(response => response.json())
      .then(data => {
        console.log('GET success:', data);
        setItems(data.items);
        onLoadCompleted && onLoadCompleted();
      })
      .catch(error => {
        console.error('GET error:', error)
      })
  }

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  return (
    <div id="item-grid">
      <ul className="item-list">
        {items ? items.map((item) => {

          const imgSource = server.concat('/image/', item.image);
          return (
            <li key={item.id} className='ItemList'>
              {/* TODO: Task 1: Replace the placeholder image with the item image */}
              {
                item.image ?
                  <img src={imgSource} /> :
                  <img src={placeholderImage} />
              }
              <p>
                <span>Name: {item.name}</span>
                <br />
                <span>Category: {item.category}</span>
              </p>
            </li>
          )
        }) : null}
      </ul>
    </div>
  )
};
