import qualified Data.ByteString
import qualified Data.Char
import Network.Socket hiding (send, sendTo, recv, recvFrom)
import Network.Socket.ByteString

newtype Foo = Foo Integer

data Blah =
    Blah1
  | Blah2 Integer;

main :: IO ()
main = do
  sock <- initSocket
  (csock, addr) <- accept sock
  bytes <- recv csock 1024
  if (toInteger (Data.ByteString.head bytes)) == (toInteger (Data.Char.ord 'a'))
    then putStrLn "got an a"
    else putStrLn "got something else"
  putStrLn $ "received " ++ (show bytes) ++ " from " ++ (show csock)

initSocket :: IO Socket
initSocket = do
  sock <- socket AF_INET Stream 0
  bind sock $ SockAddrInet 8888 0x0
  listen sock 100
  return sock
