using System.Threading;
using FoundationDB.Client;

namespace Bitgn.Layers.Tests {
    public static class Env {
        
        public static void Init(IFdbDatabase db, CancellationToken token) {
            Db = db;
            Token = token;
        }

        public static IFdbDatabase Db { get; private set; }
        public static CancellationToken Token { get; private set; }
    }
}